// Package draft — единственный владелец схемы данных драфтов в Redis.
//
// Драфт живёт в двух ключах и ZSET активных драфтов:
//
//	draft:{id}        (hash) — owner, total_size, structure (JSON фичи post)
//	draft:{id}:files  (hash) — file_id → JSON слота
//	drafts:{owner}    (ZSET) — score = время истечения
//
// Фича post создаёт/читает/удаляет драфты, фича storage меняет статусы слотов
// при загрузке — обе ТОЛЬКО через методы Store этого пакета. Формат слота,
// ключи и переходы статусов меняются здесь и нигде больше.
package draft

import (
  "encoding/json"
  "errors"
)

// Status — статус слота файла в драфте.
type Status string

const (
  Pending   Status = "pending"
  Uploading Status = "uploading"
  Done      Status = "done"
)

// Slot — заявленный слот файла. Поле Section — id секции-слота (uuid).
type Slot struct {
  Name       string `json:"name"`
  Mime       string `json:"mime"`
  Size       int64  `json:"size"`
  Section    string `json:"section"`
  Position   int    `json:"position"`
  Status     Status `json:"status"`
  UploadID   string `json:"upload_id,omitempty"`
  StorageKey string `json:"storage_key,omitempty"`
}

// Draft — содержимое draft:{id}. Structure — непрозрачный JSON фичи post
// (title + секции); этот пакет хранит его как есть, не разбирая.
type Draft struct {
  Owner     string
  TotalSize int64
  Structure json.RawMessage
}

// Владелец драфта: "user:<uuid>".
const OwnerUserPrefix = "user:"

// Поля хеша draft:{id}. Не экспортируются: снаружи пакета читать их не нужно.
const (
  ownerField     = "owner"
  totalSizeField = "total_size"
  structureField = "structure"
)

func key(id string) string      { return "draft:" + id }
func filesKey(id string) string { return "draft:" + id + ":files" }
func draftsKey(owner string) string { return "drafts:" + owner }

// Доменные ошибки драфта/слота. В HTTP-статусы их переводят хендлеры фич.
var (
  ErrDraftNotFound  = errors.New("draft: draft not found")
  ErrSlotNotFound   = errors.New("draft: slot not found")
  ErrSlotNotPending = errors.New("draft: slot not pending")
  ErrSizeMismatch   = errors.New("draft: size mismatch")
  ErrTooManyDrafts  = errors.New("draft: too many active drafts")
)
