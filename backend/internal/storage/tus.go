package storage

import (
  "errors"
  "log/slog"
  "net/http"
  "strings"

  "github.com/aws/aws-sdk-go-v2/aws"
  "github.com/aws/aws-sdk-go-v2/credentials"
  "github.com/aws/aws-sdk-go-v2/service/s3"
  "github.com/google/uuid"
  "github.com/tus/tusd/v2/pkg/handler"
  "github.com/tus/tusd/v2/pkg/memorylocker"
  "github.com/tus/tusd/v2/pkg/s3store"

  "hypersharing/internal/draft"
  "hypersharing/internal/user"
)

// newTusHandler собирает tusd (s3store) с хуками на слоты и квоту и отдаёт
// готовый http.Handler. Монтируется через http.StripPrefix(BasePath, handler).
//
// Загрузки пишут в S3 через aws-sdk-go-v2 (s3store); выдача файлов идёт через
// s3x/minio-go. Оба указывают на один и тот же MinIO. Статусы слотов меняются
// только через draft.Store.
func newTusHandler(cfg Config, drafts *draft.Store) (http.Handler, error) {
  s3c := s3.New(s3.Options{
    Region:       cfg.S3Region,
    Credentials:  credentials.NewStaticCredentialsProvider(cfg.S3AccessKeyID, cfg.S3SecretAccessKey, ""),
    BaseEndpoint: aws.String(cfg.S3EndpointURL),
    UsePathStyle: cfg.S3PathStyle,
  })

  store := s3store.New(cfg.Bucket, s3c)
  store.ObjectPrefix = cfg.ObjectPrefix
  store.MetadataObjectPrefix = cfg.MetadataPrefix

  composer := handler.NewStoreComposer()
  store.UseIn(composer)
  composer.UseLocker(memorylocker.New()) // один инстанс — in-memory локер достаточен

  // MaxSize не задаём: лимит на один файл (== лимит поста) проверяется при
  // создании драфта в фиче post, а здесь сверяется заявленный размер слота.
  return handler.NewHandler(handler.Config{
    StoreComposer:             composer,
    BasePath:                  cfg.TusBasePath,
    DisableDownload:           true, // выдача только через /f/{shortKey}
    DisableConcatenation:      true, // расширение concatenation не нужно
    PreUploadCreateCallback:   preCreate(drafts),
    PreFinishResponseCallback: finish(drafts, cfg.ObjectPrefix),
    PreUploadTerminateCallback: terminate(drafts),
  })
}

// preCreate валидирует метаданные (draftId, fileId), проверяет квоту и атомарно
// занимает слот. Возвращает upload id == fileId, чтобы ключ объекта в MinIO был
// детерминирован и известен при публикации.
func preCreate(drafts *draft.Store) func(handler.HookEvent) (handler.HTTPResponse, handler.FileInfoChanges, error) {
  return func(hook handler.HookEvent) (handler.HTTPResponse, handler.FileInfoChanges, error) {
    empty := handler.FileInfoChanges{}

    draftID, fileID, ok := uploadMeta(hook.Upload)
    if !ok {
      return handler.HTTPResponse{}, empty, handler.NewError("ERR_INVALID_METADATA", "draftId and fileId metadata are required", http.StatusBadRequest)
    }
    if hook.Upload.SizeIsDeferred {
      return handler.HTTPResponse{}, empty, handler.NewError("ERR_DEFER_LENGTH", "Upload-Defer-Length is not supported", http.StatusBadRequest)
    }

    draftOwner, err := drafts.OwnerOf(hook.Context, draftID.String())
    if err != nil {
      return handler.HTTPResponse{}, empty, tusError(err)
    }
    if requesterOwner(hook) != draftOwner {
      return handler.HTTPResponse{}, empty, handler.NewError("ERR_FORBIDDEN", "not the draft owner", http.StatusForbidden)
    }

    // Квоту здесь не пере-проверяем: она уже зарезервирована атомарно при
    // создании драфта (Reserve), а повторный Check по «сегодняшнему» ключу
    // давал бы двойной учёт и обход через полночь.
    if _, err := drafts.AcquireSlot(hook.Context, draftID.String(), fileID, hook.Upload.Size); err != nil {
      return handler.HTTPResponse{}, empty, tusError(err)
    }
    return handler.HTTPResponse{}, handler.FileInfoChanges{ID: fileID.String()}, nil
  }
}

// finish помечает слот done и записывает storage_key. Файл к этому моменту уже
// в S3, поэтому ошибку бухгалтерии не превращаем в ошибку клиента: слот
// останется uploading, публикация отклонит пост, юзер перезагрузит файл.
func finish(drafts *draft.Store, objectPrefix string) func(handler.HookEvent) (handler.HTTPResponse, error) {
  return func(hook handler.HookEvent) (handler.HTTPResponse, error) {
    draftID, fileID, ok := uploadMeta(hook.Upload)
    if !ok {
      return handler.HTTPResponse{}, handler.NewError("ERR_INVALID_METADATA", "draftId and fileId metadata are required", http.StatusInternalServerError)
    }
    if _, err := drafts.FinishSlot(hook.Context, draftID.String(), fileID, storageKey(objectPrefix, fileID.String())); err != nil {
      slog.Error("finish slot", "draft_id", draftID, "file_id", fileID, "err", err)
    }
    return handler.HTTPResponse{}, nil
  }
}

// terminate возвращает слот uploading → pending, чтобы его можно было загрузить
// заново. Для уже готового слота (done) — no-op. Прервать чужую загрузку нельзя.
func terminate(drafts *draft.Store) func(handler.HookEvent) (handler.HTTPResponse, error) {
  return func(hook handler.HookEvent) (handler.HTTPResponse, error) {
    draftID, fileID, ok := uploadMeta(hook.Upload)
    if !ok {
      return handler.HTTPResponse{}, nil
    }
    draftOwner, err := drafts.OwnerOf(hook.Context, draftID.String())
    if err != nil {
      return handler.HTTPResponse{}, nil // драфт уже удалён — защищать нечего
    }
    if requesterOwner(hook) != draftOwner {
      return handler.HTTPResponse{}, handler.NewError("ERR_FORBIDDEN", "not the draft owner", http.StatusForbidden)
    }
    if err := drafts.ReleaseSlot(hook.Context, draftID.String(), fileID); err != nil {
      slog.Error("release slot", "draft_id", draftID, "file_id", fileID, "err", err)
    }
    return handler.HTTPResponse{}, nil
  }
}

// uploadMeta достаёт draftId/fileId из Upload-Metadata и валидирует как uuid.
func uploadMeta(u handler.FileInfo) (uuid.UUID, uuid.UUID, bool) {
  draftID, err1 := uuid.Parse(u.MetaData[metaDraftID])
  fileID, err2 := uuid.Parse(u.MetaData[metaFileID])
  return draftID, fileID, err1 == nil && err2 == nil
}

// requesterOwner — владелец запрашивающего (сессия из контекста + IP). Сессия
// попадает в hook.Context, потому что tus смонтирован через OptionalAuth.
func requesterOwner(hook handler.HookEvent) string {
  u := user.UserFrom(hook.Context)
  var id *uuid.UUID
  if u != nil {
    id = &u.ID
  }
  return draft.Owner(id, draft.IPFromAddr(hook.HTTPRequest.RemoteAddr))
}

// storageKey — детерминированный ключ объекта в MinIO для слота.
func storageKey(objectPrefix, id string) string {
  prefix := strings.TrimRight(objectPrefix, "/")
  if prefix == "" {
    return id
  }
  return prefix + "/" + id
}

// tusError переводит доменные ошибки в tusd-ошибки с нужным HTTP-статусом.
func tusError(err error) error {
  switch {
  case errors.Is(err, draft.ErrDraftNotFound):
    return handler.NewError("ERR_DRAFT_NOT_FOUND", "draft not found", http.StatusNotFound)
  case errors.Is(err, draft.ErrSlotNotFound):
    return handler.NewError("ERR_SLOT_NOT_FOUND", "file slot not found", http.StatusNotFound)
  case errors.Is(err, draft.ErrSlotNotPending):
    return handler.NewError("ERR_SLOT_BUSY", "file slot is already uploading", http.StatusConflict)
  case errors.Is(err, draft.ErrSizeMismatch):
    return handler.NewError("ERR_SIZE_MISMATCH", "upload length does not match declared size", http.StatusBadRequest)
  default:
    return err
  }
}
