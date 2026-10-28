package storage

import (
  "context"
  "errors"

  "github.com/jackc/pgx/v5"

  "hypersharing/internal/storage/db"
)

// FileRepository — доступ к таблице files поверх сгенерированного sqlc.
type FileRepository struct {
  q *db.Queries
}

func NewFileRepository(dbtx db.DBTX) *FileRepository {
  return &FileRepository{q: db.New(dbtx)}
}

func (r *FileRepository) GetByShortKey(ctx context.Context, shortKey string) (File, error) {
  row, err := r.q.GetFileByShortKey(ctx, shortKey)
  if err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
      return File{}, ErrFileNotFound
    }
    return File{}, err
  }
  return File{
    ID:         row.ID,
    PostID:     row.PostID,
    Name:       row.Name,
    MimeType:   row.MimeType,
    Size:       row.Size,
    StorageKey: row.StorageKey,
  }, nil
}
