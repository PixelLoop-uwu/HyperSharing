package httpx

import "net/http"

// OriginCheck — второй слой CSRF-защиты поверх SameSite=Lax у session-cookie.
// На изменяющих запросах (POST/PUT/PATCH/DELETE) требует, чтобы заголовок Origin
// был в списке разрешённых. Отсутствие Origin (curl, не-браузер) пропускается:
// CSRF-атака идёт из браузера, а браузер Origin на такие методы всегда шлёт.
func OriginCheck(allowed []string) Middleware {
  allow := make(map[string]struct{}, len(allowed))
  for _, o := range allowed {
    allow[o] = struct{}{}
  }
  return func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      if !isUnsafeMethod(r.Method) {
        next.ServeHTTP(w, r)
        return
      }
      origin := r.Header.Get("Origin")
      if origin == "" {
        next.ServeHTTP(w, r)
        return
      }
      if _, ok := allow[origin]; !ok {
        Error(w, http.StatusForbidden, "bad_origin", "request origin not allowed")
        return
      }
      next.ServeHTTP(w, r)
    })
  }
}

func isUnsafeMethod(m string) bool {
  switch m {
  case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
    return true
  }
  return false
}
