package s3x

import (
  "context"
  "errors"
  "fmt"
  "io"
  "time"

  "github.com/minio/minio-go/v7"
)

// ErrNotFound — объекта с таким ключом нет. Проверяй через errors.Is.
var ErrNotFound = errors.New("s3x: object not found")

// unknownSizePartSize — размер части multipart при загрузке потока неизвестной длины.
// Без явного значения minio-go рассчитывает его от максимального размера объекта (5 TiB)
// и выделяет буфер в сотни мегабайт на каждую загрузку.
// Побочный эффект: потолок для потока неизвестной длины = 10000 частей * 16 MiB ≈ 160 GB.
const unknownSizePartSize = 16 << 20

// ObjectInfo — метаданные объекта без minio-типов.
type ObjectInfo struct {
  Key          string
  Size         int64
  ETag         string
  ContentType  string
  LastModified time.Time
}

// Bucket — хэндл на один бакет.
type Bucket struct {
  mc   *minio.Client
  name string
}

// Put стримит r в объект key. Если размер заранее неизвестен, передай size = -1.
// Если размер известен, передавай его: загрузка станет дешевле и проверится длина.
func (b *Bucket) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) (ObjectInfo, error) {
  opts := minio.PutObjectOptions{ContentType: contentType}
  if size < 0 {
    opts.PartSize = unknownSizePartSize
  }

  info, err := b.mc.PutObject(ctx, b.name, key, r, size, opts)
  if err != nil {
    return ObjectInfo{}, fmt.Errorf("s3x: put %s/%s: %w", b.name, key, err)
  }
  return ObjectInfo{
    Key:          info.Key,
    Size:         info.Size,
    ETag:         info.ETag,
    ContentType:  contentType,
    LastModified: info.LastModified,
  }, nil
}

// Get открывает объект на чтение. Возвращает ErrNotFound, если объекта нет.
//
// Результат умеет Seek, поэтому его можно отдавать прямо в http.ServeContent:
// Range, If-Modified-Since и докачка заработают без дополнительного кода.
// Вызывающий обязан закрыть результат.
//
// ctx живёт всё время чтения: не передавай контекст с таймаутом короче скачивания.
func (b *Bucket) Get(ctx context.Context, key string) (io.ReadSeekCloser, ObjectInfo, error) {
  obj, err := b.mc.GetObject(ctx, b.name, key, minio.GetObjectOptions{})
  if err != nil {
    return nil, ObjectInfo{}, fmt.Errorf("s3x: get %s/%s: %w", b.name, key, err)
  }
  // GetObject ленивый: несуществующий ключ вскрывается только при первом обращении.
  // Stat заодно кэширует метаданные внутри obj, что нужно для Seek(0, io.SeekEnd).
  st, err := obj.Stat()
  if err != nil {
    _ = obj.Close()
    return nil, ObjectInfo{}, fmt.Errorf("s3x: get %s/%s: %w", b.name, key, notFoundOr(err))
  }
  return obj, toInfo(st), nil
}

// Stat возвращает метаданные без скачивания тела. ErrNotFound, если объекта нет.
func (b *Bucket) Stat(ctx context.Context, key string) (ObjectInfo, error) {
  st, err := b.mc.StatObject(ctx, b.name, key, minio.StatObjectOptions{})
  if err != nil {
    return ObjectInfo{}, fmt.Errorf("s3x: stat %s/%s: %w", b.name, key, notFoundOr(err))
  }
  return toInfo(st), nil
}

// Remove удаляет объект. Идемпотентно: удаление несуществующего ключа не ошибка.
func (b *Bucket) Remove(ctx context.Context, key string) error {
  err := b.mc.RemoveObject(ctx, b.name, key, minio.RemoveObjectOptions{})
  if err != nil {
    return fmt.Errorf("s3x: remove %s/%s: %w", b.name, key, err)
  }
  return nil
}

// List перечисляет объекты с префиксом (рекурсивно).
func (b *Bucket) List(ctx context.Context, prefix string) ([]ObjectInfo, error) {
  ch := b.mc.ListObjects(ctx, b.name, minio.ListObjectsOptions{Prefix: prefix, Recursive: true})
  var out []ObjectInfo
  for obj := range ch {
    if obj.Err != nil {
      return nil, fmt.Errorf("s3x: list %s/%s: %w", b.name, prefix, obj.Err)
    }
    out = append(out, toInfo(obj))
  }
  return out, nil
}

// MultipartInfo — незавершённая multipart-загрузка.
type MultipartInfo struct {
  Key       string
  UploadID  string
  Initiated time.Time
}

// ListIncompleteUploads перечисляет незавершённые multipart-загрузки с префиксом.
func (b *Bucket) ListIncompleteUploads(ctx context.Context, prefix string) ([]MultipartInfo, error) {
  ch := b.mc.ListIncompleteUploads(ctx, b.name, prefix, true)
  var out []MultipartInfo
  for obj := range ch {
    if obj.Err != nil {
      return nil, fmt.Errorf("s3x: list uploads %s/%s: %w", b.name, prefix, obj.Err)
    }
    out = append(out, MultipartInfo{Key: obj.Key, UploadID: obj.UploadID, Initiated: obj.Initiated})
  }
  return out, nil
}

// AbortUpload отменяет незавершённую multipart-загрузку по ключу объекта.
func (b *Bucket) AbortUpload(ctx context.Context, key string) error {
  if err := b.mc.RemoveIncompleteUpload(ctx, b.name, key); err != nil {
    return fmt.Errorf("s3x: abort upload %s/%s: %w", b.name, key, err)
  }
  return nil
}

func toInfo(st minio.ObjectInfo) ObjectInfo {
  return ObjectInfo{
    Key:          st.Key,
    Size:         st.Size,
    ETag:         st.ETag,
    ContentType:  st.ContentType,
    LastModified: st.LastModified,
  }
}

// notFoundOr превращает S3-ошибку NoSuchKey в ErrNotFound, остальные отдаёт как есть.
func notFoundOr(err error) error {
  if minio.ToErrorResponse(err).Code == "NoSuchKey" {
    return ErrNotFound
  }
  return err
}