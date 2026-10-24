package user

import (
  "crypto/sha256"
  "encoding/hex"
  "errors"
  "net/http"
  "time"

  "hypersharing/internal/httpx"
  "hypersharing/internal/platform/ratelimit"
)

const cookieName = "session"

type Handler struct {
  svc             *Service
  secure          bool
  maxAge          time.Duration
  loginLimiter    *ratelimit.Limiter
  emailLoginLimit int
  window          time.Duration
}

func NewHandler(svc *Service, secure bool, maxAge time.Duration, limiter *ratelimit.Limiter, emailLoginLimit int, window time.Duration) *Handler {
  return &Handler{svc: svc, secure: secure, maxAge: maxAge, loginLimiter: limiter, emailLoginLimit: emailLoginLimit, window: window}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
  in, ok := httpx.DecodeJSON[RegisterInput](w, r)
  if !ok {
    return
  }
  u, token, err := h.svc.Register(r.Context(), in)
  if err != nil {
    h.writeError(w, r, err)
    return
  }
  h.setCookie(w, token)
  httpx.JSON(w, http.StatusCreated, u)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
  in, ok := httpx.DecodeJSON[LoginInput](w, r)
  if !ok {
    return
  }
  // Лимит попыток на аккаунт (по хешу email) — независимо от IP, против
  // распределённого перебора пароля одного пользователя.
  ok, err := h.loginLimiter.Allow(r.Context(), "rl:login_email:"+hashEmail(normalizeEmail(in.Email)), h.emailLoginLimit, h.window)
  if err != nil {
    httpx.Internal(w, r, err)
    return
  }
  if !ok {
    httpx.Error(w, http.StatusTooManyRequests, "rate_limited", "too many login attempts")
    return
  }
  u, token, err := h.svc.Login(r.Context(), in)
  if err != nil {
    h.writeError(w, r, err)
    return
  }
  h.setCookie(w, token)
  httpx.JSON(w, http.StatusOK, u)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
  if err := h.svc.Logout(r.Context(), tokenFromCookie(r)); err != nil {
    h.writeError(w, r, err)
    return
  }
  h.clearCookie(w)
  w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
  u := UserFrom(r.Context())
  if u == nil { // RequireAuth гарантирует не-nil; защита от неверного монтажа маршрута
    httpx.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
    return
  }
  httpx.JSON(w, http.StatusOK, u)
}

func (h *Handler) setCookie(w http.ResponseWriter, token string) {
  http.SetCookie(w, &http.Cookie{
    Name:     cookieName,
    Value:    token,
    Path:     "/",
    HttpOnly: true,
    Secure:   h.secure,
    SameSite: http.SameSiteLaxMode,
    MaxAge:   int(h.maxAge.Seconds()),
  })
}

func (h *Handler) clearCookie(w http.ResponseWriter) {
  http.SetCookie(w, &http.Cookie{
    Name:     cookieName,
    Value:    "",
    Path:     "/",
    HttpOnly: true,
    Secure:   h.secure,
    SameSite: http.SameSiteLaxMode,
    MaxAge:   -1,
  })
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
  var vErr *ValidationError
  switch {
  case errors.As(err, &vErr):
    httpx.Error(w, http.StatusBadRequest, "invalid_input", vErr.Msg)
  case errors.Is(err, ErrEmailTaken):
    httpx.Error(w, http.StatusConflict, "email_taken", "email already taken")
  case errors.Is(err, ErrUsernameTaken):
    httpx.Error(w, http.StatusConflict, "username_taken", "username already taken")
  case errors.Is(err, ErrInvalidCredentials):
    httpx.Error(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
  case errors.Is(err, ErrSessionNotFound):
    httpx.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
  default:
    httpx.Internal(w, r, err)
  }
}

// hashEmail — SHA-256 от нормализованного email: в ключах Redis не храним сырые
// адреса (PII).
func hashEmail(email string) string {
  sum := sha256.Sum256([]byte(email))
  return hex.EncodeToString(sum[:])
}
