package post

import (
  "errors"
  "net/http"
  "strconv"
  "time"

  "github.com/go-chi/chi/v5"
  "github.com/google/uuid"

  "hypersharing/internal/draft"
  "hypersharing/internal/httpx"
  "hypersharing/internal/platform/ratelimit"
  "hypersharing/internal/user"
)

type Handler struct {
  svc        *Service
  limiter    *ratelimit.Limiter
  draftLimit int
  window     time.Duration
}

func NewHandler(svc *Service, limiter *ratelimit.Limiter, draftLimit int, window time.Duration) *Handler {
  return &Handler{svc: svc, limiter: limiter, draftLimit: draftLimit, window: window}
}

func (h *Handler) CreateDraft(w http.ResponseWriter, r *http.Request) {
  in, ok := httpx.DecodeJSON[CreateDraftInput](w, r)
  if !ok {
    return
  }
  owner := resolveOwner(r, user.UserFrom(r.Context()))
  if ok, err := h.limiter.Allow(r.Context(), "rl:draft:"+owner, h.draftLimit, h.window); err != nil {
    httpx.Internal(w, r, err)
    return
  } else if !ok {
    httpx.Error(w, http.StatusTooManyRequests, "rate_limited", "too many requests")
    return
  }
  d, err := h.svc.CreateDraft(r.Context(), owner, in)
  if err != nil {
    h.writeError(w, r, err)
    return
  }
  httpx.JSON(w, http.StatusCreated, d)
}

func (h *Handler) CancelDraft(w http.ResponseWriter, r *http.Request) {
  if err := h.svc.CancelDraft(r.Context(), chi.URLParam(r, "id"), resolveOwner(r, user.UserFrom(r.Context()))); err != nil {
    h.writeError(w, r, err)
    return
  }
  w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Publish(w http.ResponseWriter, r *http.Request) {
  p, err := h.svc.Publish(r.Context(), chi.URLParam(r, "id"), resolveOwner(r, user.UserFrom(r.Context())))
  if err != nil {
    h.writeError(w, r, err)
    return
  }
  httpx.JSON(w, http.StatusOK, p)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
  id, err := uuid.Parse(chi.URLParam(r, "id"))
  if err != nil {
    httpx.Error(w, http.StatusNotFound, "not_found", "post not found")
    return
  }
  p, err := h.svc.Get(r.Context(), id)
  if err != nil {
    h.writeError(w, r, err)
    return
  }
  httpx.JSON(w, http.StatusOK, p)
}

// ListMine — GET /me/posts: свои посты, курсорная пагинация (RequireAuth).
func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
  u := user.UserFrom(r.Context())
  if u == nil { // RequireAuth гарантирует не-nil
    httpx.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
    return
  }
  out, err := h.svc.ListMine(r.Context(), u.ID, parseLimit(r), r.URL.Query().Get("cursor"))
  if err != nil {
    h.writeError(w, r, err)
    return
  }
  httpx.JSON(w, http.StatusOK, out)
}

// Quota — GET /me/quota: лимиты квоты для профиля (RequireAuth).
func (h *Handler) Quota(w http.ResponseWriter, r *http.Request) {
  u := user.UserFrom(r.Context())
  if u == nil { // RequireAuth гарантирует не-nil; защита от неверного монтажа
    httpx.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
    return
  }
  q, err := h.svc.Quota(r.Context(), u.ID)
  if err != nil {
    httpx.Internal(w, r, err)
    return
  }
  httpx.JSON(w, http.StatusOK, q)
}

func parseLimit(r *http.Request) int {
  const defaultLimit, maxLimit = 20, 100
  s := r.URL.Query().Get("limit")
  if s == "" {
    return defaultLimit
  }
  n, err := strconv.Atoi(s)
  if err != nil || n < 1 {
    return defaultLimit
  }
  if n > maxLimit {
    return maxLimit
  }
  return n
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
  id, err := uuid.Parse(chi.URLParam(r, "id"))
  if err != nil {
    httpx.Error(w, http.StatusNotFound, "not_found", "post not found")
    return
  }
  u := user.UserFrom(r.Context())
  if u == nil { // RequireAuth гарантирует не-nil; защита от неверного монтажа
    httpx.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
    return
  }
  if err := h.svc.Delete(r.Context(), id, u.ID); err != nil {
    h.writeError(w, r, err)
    return
  }
  w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
  var vErr *ValidationError
  switch {
  case errors.As(err, &vErr):
    httpx.Error(w, http.StatusBadRequest, "invalid_input", vErr.Msg)
  case errors.Is(err, draft.ErrDraftNotFound), errors.Is(err, ErrPostNotFound):
    httpx.Error(w, http.StatusNotFound, "not_found", "not found")
  case errors.Is(err, ErrNotAllUploaded):
    httpx.Error(w, http.StatusConflict, "not_all_uploaded", "not all files uploaded yet")
  case errors.Is(err, ErrQuotaExceeded):
    httpx.Error(w, http.StatusTooManyRequests, "quota_exceeded", "daily upload quota exceeded")
  case errors.Is(err, ErrStorageQuotaExceeded):
    httpx.Error(w, http.StatusRequestEntityTooLarge, "storage_quota_exceeded", "storage quota exceeded")
  case errors.Is(err, draft.ErrTooManyDrafts):
    httpx.Error(w, http.StatusTooManyRequests, "too_many_drafts", "too many active drafts")
  case errors.Is(err, ErrAlreadyPublishing):
    httpx.Error(w, http.StatusConflict, "publishing", "publish already in progress")
  case errors.Is(err, ErrForbidden):
    httpx.Error(w, http.StatusForbidden, "forbidden", "forbidden")
  default:
    httpx.Internal(w, r, err)
  }
}
