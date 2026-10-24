package user

import (
  "context"
  "net/http"

  "hypersharing/internal/httpx"
)

type ctxKey struct{}

func WithUser(ctx context.Context, u *User) context.Context {
  return context.WithValue(ctx, ctxKey{}, u)
}

// UserFrom возвращает текущего пользователя или nil для анонима.
func UserFrom(ctx context.Context) *User {
  u, _ := ctx.Value(ctxKey{}).(*User)
  return u
}

// RequireAuth пропускает дальше только с действующей сессией, иначе 401.
func RequireAuth(svc *Service) func(http.Handler) http.Handler {
  return func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      u, err := svc.ResolveSession(r.Context(), tokenFromCookie(r))
      if err != nil {
        httpx.Internal(w, r, err)
        return
      }
      if u == nil {
        httpx.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
        return
      }
      next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), u)))
    })
  }
}

// OptionalAuth кладёт пользователя в контекст при действующей сессии, иначе аноним (nil).
func OptionalAuth(svc *Service) func(http.Handler) http.Handler {
  return func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      u, err := svc.ResolveSession(r.Context(), tokenFromCookie(r))
      if err != nil {
        httpx.Internal(w, r, err)
        return
      }
      next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), u)))
    })
  }
}

func tokenFromCookie(r *http.Request) string {
  c, err := r.Cookie(cookieName)
  if err != nil {
    return ""
  }
  return c.Value
}
