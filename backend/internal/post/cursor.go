package post

import (
  "encoding/base64"
  "fmt"
  "strconv"
  "strings"
  "time"

  "github.com/google/uuid"
)

// Курсор для списка постов — opaque-токен, кодирующий (created_at, id) последнего
// элемента. Ключом служит строго «старше» в DESC-порядке (keyset pagination),
// поэтому вставка новых постов между страницами не сдвигает и не дублирует их.

func encodeCursor(t time.Time, id uuid.UUID) string {
  raw := fmt.Sprintf("%d|%s", t.UnixNano(), id.String())
  return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(s string) (time.Time, uuid.UUID, error) {
  raw, err := base64.RawURLEncoding.DecodeString(s)
  if err != nil {
    return time.Time{}, uuid.Nil, fmt.Errorf("decode cursor: %w", err)
  }
  parts := strings.SplitN(string(raw), "|", 2)
  if len(parts) != 2 {
    return time.Time{}, uuid.Nil, fmt.Errorf("decode cursor: bad format")
  }
  nano, err := strconv.ParseInt(parts[0], 10, 64)
  if err != nil {
    return time.Time{}, uuid.Nil, fmt.Errorf("decode cursor: %w", err)
  }
  id, err := uuid.Parse(parts[1])
  if err != nil {
    return time.Time{}, uuid.Nil, fmt.Errorf("decode cursor: %w", err)
  }
  return time.Unix(0, nano), id, nil
}
