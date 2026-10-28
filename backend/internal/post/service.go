package post

import (
  "context"
  "encoding/json"
  "errors"
  "log/slog"
  "sort"
  "strings"
  "time"

  "github.com/google/uuid"
  "github.com/jackc/pgx/v5/pgconn"
  "github.com/jackc/pgx/v5/pgxpool"
  "github.com/redis/go-redis/v9"

  "hypersharing/internal/draft"
  "hypersharing/internal/storage"
)

const publishLockTTL = 30 * time.Second

// Service — бизнес-логика постов: драфты, публикация, чтение, удаление.
// Схема драфта (ключи, слоты, статусы) живёт в пакете draft; квоту резервирует
// storage. post импортирует storage только за хранилищем квоты.
type Service struct {
  pool      *pgxpool.Pool
  repo      *Repository
  drafts    *draft.Store
  quota     *storage.Quota
  rdb       *redis.Client
  limits    Limits
  apiPrefix string // "/api/v1" — префикс публичных URL файлов и аватаров
}

func New(pool *pgxpool.Pool, rdb *redis.Client, quota *storage.Quota, drafts *draft.Store, limits Limits, apiPrefix string) *Service {
  return &Service{
    pool:      pool,
    repo:      NewRepository(pool),
    drafts:    drafts,
    quota:     quota,
    rdb:       rdb,
    limits:    limits,
    apiPrefix: apiPrefix,
  }
}

// CreateDraft валидирует структуру поста, проверяет лимиты и квоту, резервирует
// суточную квоту и кладёт драфт в Redis. Возвращает id драфта и id слотов.
func (s *Service) CreateDraft(ctx context.Context, userID uuid.UUID, in CreateDraftInput) (DraftRead, error) {
  totalSize, err := validateDraft(s.limits, in)
  if err != nil {
    return DraftRead{}, err
  }

  used, err := s.repo.UserStorageUsed(ctx, userID)
  if err != nil {
    return DraftRead{}, err
  }
  if used+totalSize > s.limits.StorageQuota {
    return DraftRead{}, ErrStorageQuotaExceeded
  }

  owner := draft.Owner(userID)
  // Резерв суточной квоты. Брошенный драфт резерв не возвращает.
  if err := s.quota.Reserve(ctx, owner, totalSize); err != nil {
    if errors.Is(err, storage.ErrQuotaExceeded) {
      return DraftRead{}, ErrQuotaExceeded
    }
    return DraftRead{}, err
  }

  id := uuid.New()
  structure := draftStructure{Title: in.Title, Sections: make([]draftSection, 0, len(in.Sections))}
  slots := make(map[string]draft.Slot)
  resp := DraftRead{ID: id.String(), Sections: make([]DraftSectionRead, 0, len(in.Sections))}

  for _, sec := range in.Sections {
    secID := uuid.New()
    structure.Sections = append(structure.Sections, draftSection{ID: secID.String(), Body: sec.Body})
    dsr := DraftSectionRead{ID: secID.String(), Files: make([]DraftFileRead, 0, len(sec.Files))}
    for i, f := range sec.Files {
      fileID := uuid.New()
      slots[fileID.String()] = draft.Slot{
        Name:     f.Name,
        Mime:     f.Mime,
        Size:     f.Size,
        Section:  secID.String(),
        Position: i,
        Status:   draft.Pending,
      }
      dsr.Files = append(dsr.Files, DraftFileRead{ID: fileID.String()})
    }
    resp.Sections = append(resp.Sections, dsr)
  }

  raw, err := json.Marshal(structure)
  if err != nil {
    return DraftRead{}, err
  }
  if err := s.drafts.Create(ctx, id.String(), owner, totalSize, raw, slots, s.limits.MaxDrafts); err != nil {
    // Драфт не создался — возвращаем резерв квоты, чтобы не «съесть» лимит зря.
    if rerr := s.quota.Release(ctx, owner, totalSize); rerr != nil {
      slog.Error("release quota after failed draft", "owner", owner, "err", rerr)
    }
    return DraftRead{}, err
  }
  return resp, nil
}

// CancelDraft удаляет драфт и возвращает резерв квоты. Только владелец драфта.
// Брошенные tus-загрузки в S3 приберёт фоновый воркер.
func (s *Service) CancelDraft(ctx context.Context, id string, requesterOwner string) error {
  d, err := s.drafts.Get(ctx, id)
  if err != nil {
    return err
  }
  if d.Owner != requesterOwner {
    // Не раскрываем, что драфт существует: чужой и отсутствующий — одинаково 404.
    return draft.ErrDraftNotFound
  }
  // Возвращаем квоту только за слоты, в которые ничего не загружалось (pending).
  // За uploading/done байты уже лежат в S3 — их возврат дал бы обход суточной квоты.
  slots, err := s.drafts.Slots(ctx, id)
  if err != nil {
    return err
  }
  var refund int64
  for _, slot := range slots {
    if slot.Status == draft.Pending {
      refund += slot.Size
    }
  }
  if err := s.drafts.Delete(ctx, d.Owner, id); err != nil {
    return err
  }
  if refund > 0 {
    if err := s.quota.Release(ctx, d.Owner, refund); err != nil {
      // Резерв останется зарезервированным (безопасное направление) — не блокируем отмену.
      slog.Error("release quota on draft cancel", "draft", id, "err", err)
    }
  }
  return nil
}

// Publish переносит драфт из Redis в Postgres. Только владелец драфта.
// Идемпотентен: повторный вызов после успеха возвращает уже опубликованный пост.
func (s *Service) Publish(ctx context.Context, id string, requesterOwner string) (PostRead, error) {
  draftID, err := uuid.Parse(id)
  if err != nil {
    return PostRead{}, draft.ErrDraftNotFound
  }

  // Защита от двойного клика. Настоящая гарантия уникальности — PK posts.id.
  lockKey := "publish:" + id
  ok, err := s.rdb.SetNX(ctx, lockKey, "1", publishLockTTL).Result()
  if err != nil {
    return PostRead{}, err
  }
  if !ok {
    return PostRead{}, ErrAlreadyPublishing
  }
  defer func() { _ = s.rdb.Del(ctx, lockKey).Err() }()

  // Идемпотентность: если пост уже есть (повторный publish после успеха, или
  // конкурентная публикация завершилась раньше) — возвращаем существующий.
  // Опубликованный пост публичен, поэтому здесь владелец не проверяется.
  if p, err := s.Get(ctx, draftID); err == nil {
    return p, nil
  } else if !errors.Is(err, ErrPostNotFound) {
    return PostRead{}, err
  }

  d, err := s.drafts.Get(ctx, id)
  if err != nil {
    return PostRead{}, err
  }
  if d.Owner != requesterOwner {
    // Чужой драфт неотличим от отсутствующего: 404, чтобы не подтверждать uuid.
    return PostRead{}, draft.ErrDraftNotFound
  }
  var structure draftStructure
  if err := json.Unmarshal(d.Structure, &structure); err != nil {
    return PostRead{}, err
  }
  slots, err := s.drafts.Slots(ctx, id)
  if err != nil {
    return PostRead{}, err
  }
  for _, slot := range slots {
    if slot.Status != draft.Done {
      return PostRead{}, ErrNotAllUploaded
    }
  }

  authorID, err := ownerUserID(d.Owner)
  if err != nil {
    return PostRead{}, err
  }

  // short_key 62^8 — коллизия крайне редка, но на UNIQUE-нарушении перегенерируем
  // ключи и пробуем транзакцию снова, чтобы публикация не падала 500-кой.
  const maxShortKeyRetries = 5
  for attempt := 0; ; attempt++ {
    err := s.persistPost(ctx, draftID, authorID, structure, slots, d.TotalSize)
    if err == nil {
      break
    }
    if !isShortKeyCollision(err) || attempt >= maxShortKeyRetries-1 {
      return PostRead{}, err
    }
  }

  // Ключи драфта больше не нужны. Неудача здесь безопасна: TTL их добьёт.
  if err := s.drafts.Delete(ctx, d.Owner, id); err != nil {
    slog.Error("delete draft after publish", "draft", id, "err", err)
  }

  return s.Get(ctx, draftID)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (PostRead, error) {
  p, err := s.repo.GetPost(ctx, id)
  if err != nil {
    return PostRead{}, err
  }
  return s.buildRead(ctx, p)
}

// GetByShortKey читает пост по публичному ключу из URL /{short_key}.
func (s *Service) GetByShortKey(ctx context.Context, shortKey string) (PostRead, error) {
  p, err := s.repo.GetPostByShortKey(ctx, shortKey)
  if err != nil {
    return PostRead{}, err
  }
  return s.buildRead(ctx, p)
}

func (s *Service) buildRead(ctx context.Context, p Post) (PostRead, error) {
  sections, err := s.repo.Sections(ctx, p.ID)
  if err != nil {
    return PostRead{}, err
  }
  files, err := s.repo.Files(ctx, p.ID)
  if err != nil {
    return PostRead{}, err
  }
  bySection := make(map[uuid.UUID][]File, len(sections))
  for _, f := range files {
    bySection[f.SectionID] = append(bySection[f.SectionID], f)
  }

  out := PostRead{
    ID:        p.ID,
    ShortKey:  p.ShortKey,
    Title:     p.Title,
    CreatedAt: p.CreatedAt,
    Sections:  make([]SectionRead, 0, len(sections)),
  }
  author := &AuthorRead{Username: p.AuthorUsername}
  if p.AvatarVersion != "" {
    url := "/avatars/" + p.AuthorID.String() + "?v=" + p.AvatarVersion
    author.AvatarURL = &url
  }
  out.Author = author
  for _, sec := range sections {
    sr := SectionRead{Body: sec.Body, Files: make([]FileRead, 0, len(bySection[sec.ID]))}
    for _, f := range bySection[sec.ID] {
      sr.Files = append(sr.Files, FileRead{Name: f.Name, Mime: f.MimeType, Size: f.Size, URL: "/f/" + f.ShortKey})
    }
    out.Sections = append(out.Sections, sr)
  }
  return out, nil
}

// ListMine — курсорная пагинация постов автора (новые первыми). limit берётся с
// запасом +1, чтобы понять, есть ли следующая страница.
func (s *Service) ListMine(ctx context.Context, userID uuid.UUID, limit int, cursor string) (PostListRead, error) {
  var (
    items []PostListItem
    err   error
  )
  if cursor == "" {
    items, err = s.repo.ListByAuthor(ctx, userID, limit+1)
  } else {
    createdAt, cursorID, derr := decodeCursor(cursor)
    if derr != nil {
      return PostListRead{}, &ValidationError{"invalid cursor"}
    }
    items, err = s.repo.ListByAuthorAfter(ctx, userID, createdAt, cursorID, limit+1)
  }
  if err != nil {
    return PostListRead{}, err
  }

  out := PostListRead{Posts: items}
  if len(items) > limit {
    out.Posts = items[:limit]
    last := items[limit-1]
    next := encodeCursor(last.CreatedAt, last.ID)
    out.NextCursor = &next
  }
  return out, nil
}

// Quota возвращает текущие лимиты пользователя для профиля: занятое хранилище и
// суточный лимит загрузки. Значения условные: Daily.Used — это зарезервированные
// байты (включая ещё недозагруженные драфты), а не фактически залитое.
func (s *Service) Quota(ctx context.Context, userID uuid.UUID) (QuotaRead, error) {
  owner := draft.Owner(userID)
  usedStorage, err := s.repo.UserStorageUsed(ctx, userID)
  if err != nil {
    return QuotaRead{}, err
  }
  usedDaily, err := s.quota.Get(ctx, owner)
  if err != nil {
    return QuotaRead{}, err
  }
  return QuotaRead{
    Storage: QuotaMetric{Used: usedStorage, Limit: s.limits.StorageQuota},
    Daily:   QuotaMetric{Used: usedDaily, Limit: s.quota.Limit()},
  }, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
  p, err := s.repo.GetPost(ctx, id)
  if err != nil {
    return err
  }
  if p.AuthorID != userID {
    return ErrForbidden
  }
  // Объекты в MinIO удаляет воркер (транзакция БД до S3 не дотягивается).
  return s.repo.DeletePost(ctx, id)
}

// persistPost создаёт пост, секции и файлы в одной транзакции. short_key
// генерируются внутри — при UNIQUE-нарушении вызывающий пробует снова.
func (s *Service) persistPost(ctx context.Context, draftID uuid.UUID, authorID uuid.UUID, structure draftStructure, slots map[string]draft.Slot, totalSize int64) error {
  tx, err := s.pool.Begin(ctx)
  if err != nil {
    return err
  }
  defer func() { _ = tx.Rollback(ctx) }()
  txn := s.repo.WithTx(tx)

  postShortKey, err := newShortKey()
  if err != nil {
    return err
  }
  if err := txn.CreatePost(ctx, draftID, authorID, postShortKey, structure.Title, totalSize); err != nil {
    return err
  }
  for i, sec := range structure.Sections {
    secID, err := uuid.Parse(sec.ID)
    if err != nil {
      return err
    }
    if err := txn.CreateSection(ctx, secID, draftID, i, sec.Body); err != nil {
      return err
    }
    for _, sf := range filesOfSection(slots, sec.ID) {
      fileID, err := uuid.Parse(sf.fileID)
      if err != nil {
        return err
      }
      shortKey, err := newShortKey()
      if err != nil {
        return err
      }
      if err := txn.CreateFile(ctx, fileID, draftID, secID, sf.slot.Position, sf.slot.Name, sf.slot.Mime, sf.slot.Size, sf.slot.StorageKey, shortKey); err != nil {
        return err
      }
    }
  }
  return tx.Commit(ctx)
}

// isShortKeyCollision отличает UNIQUE-нарушение по short_key (поста или файла) от
// остальных ошибок, чтобы публикацию можно было безопасно повторить с новыми ключами.
func isShortKeyCollision(err error) bool {
  var pgErr *pgconn.PgError
  if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
    return false
  }
  return pgErr.ConstraintName == "files_short_key_key" || pgErr.ConstraintName == "posts_short_key_key"
}

type sectionFile struct {
  fileID string
  slot   draft.Slot
}

// filesOfSection собирает слоты секции и упорядочивает по position.
func filesOfSection(slots map[string]draft.Slot, sectionID string) []sectionFile {
  var out []sectionFile
  for fileID, slot := range slots {
    if slot.Section == sectionID {
      out = append(out, sectionFile{fileID: fileID, slot: slot})
    }
  }
  sort.Slice(out, func(i, j int) bool { return out[i].slot.Position < out[j].slot.Position })
  return out
}

func validateDraft(lim Limits, in CreateDraftInput) (int64, error) {
  if len(in.Sections) == 0 {
    return 0, &ValidationError{"post must have at least one section"}
  }
  if len(in.Title) > lim.TitleMaxLen {
    return 0, &ValidationError{"title too long"}
  }
  if len(in.Sections) > lim.MaxSections {
    return 0, &ValidationError{"too many sections"}
  }

  maxFiles := lim.MaxFiles
  maxBytes := lim.MaxBytes
  fileCount := 0
  var totalSize int64
  for _, sec := range in.Sections {
    if len(sec.Body) > lim.SectionMaxLen {
      return 0, &ValidationError{"section text too long"}
    }
    fileCount += len(sec.Files)
    for _, f := range sec.Files {
      if f.Size <= 0 {
        return 0, &ValidationError{"file size must be positive"}
      }
      if f.Size > maxBytes {
        return 0, &ValidationError{"file too large"}
      }
      if f.Name == "" || len(f.Name) > 255 {
        return 0, &ValidationError{"invalid file name"}
      }
      // mime уходит в Content-Type при раздаче — не пускаем CR/LF (header-инъекция).
      if len(f.Mime) > 128 || strings.ContainsAny(f.Mime, "\r\n") {
        return 0, &ValidationError{"invalid mime type"}
      }
      totalSize += f.Size
    }
  }
  if fileCount > maxFiles {
    return 0, &ValidationError{"too many files"}
  }
  if totalSize > maxBytes {
    return 0, &ValidationError{"post too large"}
  }
  return totalSize, nil
}
