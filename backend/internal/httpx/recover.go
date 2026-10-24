package httpx

import (
  "net/http"
  "runtime/debug"
)

func Recover(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    defer func() {
      rec := recover()
      if rec == nil {
        return
      }
      if rec == http.ErrAbortHandler {
        panic(rec) // intentional abort by net/http, pass it on
      }
      Logger(r.Context()).Error("panic", "panic", rec, "stack", string(debug.Stack()))
      Error(w, http.StatusInternalServerError, "internal", "internal error")
    }()
    next.ServeHTTP(w, r)
  })
}