package post

import (
  "fmt"
  "strings"

  "github.com/google/uuid"

  "hypersharing/internal/draft"
)

// ownerUserID достаёт uuid пользователя из owner-строки "user:<uuid>".
func ownerUserID(owner string) (uuid.UUID, error) {
  if !strings.HasPrefix(owner, draft.OwnerUserPrefix) {
    return uuid.Nil, fmt.Errorf("owner %q is not a user", owner)
  }
  return uuid.Parse(owner[len(draft.OwnerUserPrefix):])
}
