package post

import (
  "context"
  "errors"
  "time"

  "github.com/google/uuid"
  "github.com/jackc/pgx/v5"
  "github.com/jackc/pgx/v5/pgtype"

  "hypersharing/internal/post/db"
)

// Repository — доступ к постам поверх сгенерированного sqlc. Создание поста
// идёт в транзакции: сервис открывает tx и получает Repository, привязанный к
// нему через WithTx.
type Repository struct {
  q *db.Queries
}

func NewRepository(dbtx db.DBTX) *Repository {
  return &Repository{q: db.New(dbtx)}
}

func (r *Repository) WithTx(tx pgx.Tx) *Repository {
  return &Repository{q: r.q.WithTx(tx)}
}

func (r *Repository) CreatePost(ctx context.Context, id uuid.UUID, authorID uuid.UUID, shortKey, title string, totalSize int64) error {
  _, err := r.q.CreatePost(ctx, db.CreatePostParams{
    ID:        id,
    AuthorID:  authorID,
    ShortKey:  shortKey,
    Title:     title,
    TotalSize: totalSize,
  })
  return err
}

func (r *Repository) CreateSection(ctx context.Context, id, postID uuid.UUID, position int, body string) error {
  _, err := r.q.CreateSection(ctx, db.CreateSectionParams{
    ID:       id,
    PostID:   postID,
    Position: int32(position),
    Body:     body,
  })
  return err
}

func (r *Repository) CreateFile(ctx context.Context, id, postID, sectionID uuid.UUID, position int, name, mimeType string, size int64, storageKey, shortKey string) error {
  _, err := r.q.CreateFile(ctx, db.CreateFileParams{
    ID:         id,
    PostID:     postID,
    SectionID:  sectionID,
    Position:   int32(position),
    Name:       name,
    MimeType:   mimeType,
    Size:       size,
    StorageKey: storageKey,
    ShortKey:   shortKey,
  })
  return err
}

func (r *Repository) GetPost(ctx context.Context, id uuid.UUID) (Post, error) {
  row, err := r.q.GetPost(ctx, id)
  if err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
      return Post{}, ErrPostNotFound
    }
    return Post{}, err
  }
  return postFromRow(row.ID, row.AuthorID, row.ShortKey, row.Title, row.TotalSize, row.CreatedAt, row.Username, row.AvatarVersion), nil
}

func (r *Repository) GetPostByShortKey(ctx context.Context, shortKey string) (Post, error) {
  row, err := r.q.GetPostByShortKey(ctx, shortKey)
  if err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
      return Post{}, ErrPostNotFound
    }
    return Post{}, err
  }
  return postFromRow(row.ID, row.AuthorID, row.ShortKey, row.Title, row.TotalSize, row.CreatedAt, row.Username, row.AvatarVersion), nil
}

// postFromRow собирает доменную модель Post из строки sqlc. GetPostRow и
// GetPostByShortKeyRow имеют одинаковый набор полей, поэтому маппинг общий.
func postFromRow(id, authorID uuid.UUID, shortKey, title string, totalSize int64, createdAt time.Time, username, avatarVersion pgtype.Text) Post {
  p := Post{
    ID:        id,
    AuthorID:  authorID,
    ShortKey:  shortKey,
    Title:     title,
    TotalSize: totalSize,
    CreatedAt: createdAt,
  }
  if username.Valid {
    p.AuthorUsername = username.String
  }
  if avatarVersion.Valid {
    p.AvatarVersion = avatarVersion.String
  }
  return p
}

func (r *Repository) Sections(ctx context.Context, postID uuid.UUID) ([]Section, error) {
  rows, err := r.q.GetSectionsByPost(ctx, postID)
  if err != nil {
    return nil, err
  }
  sections := make([]Section, len(rows))
  for i, row := range rows {
    sections[i] = Section{ID: row.ID, Position: int(row.Position), Body: row.Body}
  }
  return sections, nil
}

func (r *Repository) Files(ctx context.Context, postID uuid.UUID) ([]File, error) {
  rows, err := r.q.GetFilesByPost(ctx, postID)
  if err != nil {
    return nil, err
  }
  files := make([]File, len(rows))
  for i, row := range rows {
    files[i] = File{
      SectionID: row.SectionID,
      Position:  int(row.Position),
      Name:      row.Name,
      MimeType:  row.MimeType,
      Size:      row.Size,
      ShortKey:  row.ShortKey,
    }
  }
  return files, nil
}

func (r *Repository) DeletePost(ctx context.Context, id uuid.UUID) error {
  _, err := r.q.DeletePost(ctx, id)
  if err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
      return ErrPostNotFound
    }
    return err
  }
  return nil
}

func (r *Repository) UserStorageUsed(ctx context.Context, authorID uuid.UUID) (int64, error) {
  return r.q.GetUserStorageUsed(ctx, authorID)
}

// AllStorageKeys возвращает все storage_key опубликованных файлов. Нужен
// воркеру для GC сиротских объектов в S3.
func (r *Repository) AllStorageKeys(ctx context.Context) ([]string, error) {
  return r.q.AllStorageKeys(ctx)
}

// ListByAuthor — страница постов автора (пост-записи в DESC). limit+1 вызывается
// сервисом, чтобы понять, есть ли следующая страница.
func (r *Repository) ListByAuthor(ctx context.Context, authorID uuid.UUID, limit int) ([]PostListItem, error) {
  rows, err := r.q.ListPostsByAuthor(ctx, db.ListPostsByAuthorParams{
    AuthorID:  authorID,
    PageLimit: int32(limit),
  })
  if err != nil {
    return nil, err
  }
  items := make([]PostListItem, len(rows))
  for i, row := range rows {
    items[i] = listItem(row.ID, row.ShortKey, row.Title, row.TotalSize, row.CreatedAt)
  }
  return items, nil
}

// ListByAuthorAfter — следующая страница после курсора (created_at, id).
func (r *Repository) ListByAuthorAfter(ctx context.Context, authorID uuid.UUID, createdAt time.Time, cursorID uuid.UUID, limit int) ([]PostListItem, error) {
  rows, err := r.q.ListPostsByAuthorAfter(ctx, db.ListPostsByAuthorAfterParams{
    AuthorID:        authorID,
    CursorCreatedAt: createdAt,
    CursorID:        cursorID,
    PageLimit:       int32(limit),
  })
  if err != nil {
    return nil, err
  }
  items := make([]PostListItem, len(rows))
  for i, row := range rows {
    items[i] = listItem(row.ID, row.ShortKey, row.Title, row.TotalSize, row.CreatedAt)
  }
  return items, nil
}

func listItem(id uuid.UUID, shortKey, title string, totalSize int64, createdAt time.Time) PostListItem {
  return PostListItem{ID: id, ShortKey: shortKey, Title: title, TotalSize: totalSize, CreatedAt: createdAt}
}
