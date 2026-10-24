package post

import (
  "net/http"
  "strings"

  "github.com/google/uuid"

  "hypersharing/internal/draft"
  "hypersharing/internal/user"
)

// resolveOwner возвращает владельца драфта для запроса: залогиненный —
// "user:<uuid>", аноним — "ip:<hash>". Формат строится в пакете draft.
func resolveOwner(r *http.Request, u *user.User) string {
  var id *uuid.UUID
  if u != nil {
    id = &u.ID
  }
  return draft.Owner(id, draft.IPFromAddr(r.RemoteAddr))
}

// ownerUserID достаёт uuid пользователя из owner, если тот залогиненный.
func ownerUserID(owner string) (uuid.UUID, bool) {
  if !strings.HasPrefix(owner, draft.OwnerUserPrefix) {
    return uuid.Nil, false
  }
  id, err := uuid.Parse(owner[len(draft.OwnerUserPrefix):])
  return id, err == nil
}
