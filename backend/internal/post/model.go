package post

import (
  "errors"
  "time"

  "github.com/google/uuid"
)

// draftStructure — JSON в поле structure хеша draft:{id}. Формат фичи post:
// этот пакет маршалит его в непрозрачный json.RawMessage для draft.Store.
type draftStructure struct {
  Title    string         `json:"title"`
  Sections []draftSection `json:"sections"`
}

type draftSection struct {
  ID   string `json:"id"`
  Body string `json:"body"`
}

// --- Входные данные ---

type CreateDraftInput struct {
  Title    string         `json:"title"`
  Sections []SectionInput `json:"sections"`
}

type SectionInput struct {
  Body  string      `json:"body"`
  Files []FileInput `json:"files"`
}

type FileInput struct {
  Name string `json:"name"`
  Mime string `json:"mime"`
  Size int64  `json:"size"`
}

// --- Ответы ---

type DraftRead struct {
  ID       string             `json:"id"`
  Sections []DraftSectionRead `json:"sections"`
}

type DraftSectionRead struct {
  ID    string          `json:"id"`
  Files []DraftFileRead `json:"files"`
}

type DraftFileRead struct {
  ID string `json:"id"`
}

type PostRead struct {
  ID        uuid.UUID     `json:"id"`
  Title     string        `json:"title"`
  Author    *AuthorRead   `json:"author"` // nil = аноним
  CreatedAt time.Time     `json:"created_at"`
  Sections  []SectionRead `json:"sections"`
}

type AuthorRead struct {
  Username  string  `json:"username"`
  AvatarURL *string `json:"avatar_url"` // nil = нет аватара
}

type SectionRead struct {
  Body  string     `json:"body"`
  Files []FileRead `json:"files"`
}

type FileRead struct {
  Name string `json:"name"`
  Mime string `json:"mime"`
  Size int64  `json:"size"`
  URL  string `json:"url"` // /f/{short_key}
}

type PostListItem struct {
  ID        uuid.UUID `json:"id"`
  Title     string    `json:"title"`
  TotalSize int64     `json:"total_size"`
  CreatedAt time.Time `json:"created_at"`
}

type PostListRead struct {
  Posts      []PostListItem `json:"posts"`
  NextCursor *string        `json:"next_cursor"` // nil = конец
}

// QuotaMetric — «использовано/лимит» для одного счётчика квоты.
type QuotaMetric struct {
  Used  int64 `json:"used"`
  Limit int64 `json:"limit"`
}

// QuotaRead — ответ GET /me/quota: занятое хранилище и суточный лимит загрузки.
type QuotaRead struct {
  Storage QuotaMetric `json:"storage"`
  Daily   QuotaMetric `json:"daily"`
}

// --- Доменные модели из БД ---

type Post struct {
  ID             uuid.UUID
  AuthorID       *uuid.UUID
  AuthorUsername string
  AvatarVersion  string
  Title          string
  TotalSize      int64
  CreatedAt      time.Time
}

type Section struct {
  ID       uuid.UUID
  Position int
  Body     string
}

type File struct {
  SectionID uuid.UUID
  Position  int
  Name      string
  MimeType  string
  Size      int64
  ShortKey  string
}

// ValidationError — невалидные входные данные. Handler отвечает на неё 400.
type ValidationError struct {
  Msg string
}

func (e *ValidationError) Error() string { return e.Msg }

// Доменные ошибки фичи post. Ошибки драфта/слота живут в пакете draft.
var (
  ErrPostNotFound         = errors.New("post: post not found")
  ErrNotAllUploaded       = errors.New("post: not all files uploaded")
  ErrQuotaExceeded        = errors.New("post: daily quota exceeded")
  ErrStorageQuotaExceeded = errors.New("post: storage quota exceeded")
  ErrAlreadyPublishing    = errors.New("post: publish already in progress")
  ErrForbidden            = errors.New("post: forbidden")
)
