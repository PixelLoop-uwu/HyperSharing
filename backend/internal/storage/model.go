package storage

import (
  "errors"

  "github.com/google/uuid"
)

// Ключи Upload-Metadata, которые клиент обязан передавать при создании
// tus-загрузки: draftId и fileId (значения base64 внутри заголовка).
const (
  metaDraftID = "draftId"
  metaFileID  = "fileId"
)

// File — строка таблицы files, как её видит выдача по short_key.
type File struct {
  ID         uuid.UUID
  PostID     uuid.UUID
  Name       string
  MimeType   string
  Size       int64
  StorageKey string
}

// Доменные ошибки storage.
var (
  ErrQuotaExceeded = errors.New("storage: quota exceeded")
  ErrFileNotFound  = errors.New("storage: file not found")
)
