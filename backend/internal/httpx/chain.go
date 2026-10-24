package httpx

import "net/http"

type Middleware func(http.Handler) http.Handler

// Chain wraps h in the given middleware. The first one listed is the
// outermost: it sees the request first and the response last.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
  for i := len(mws) - 1; i >= 0; i-- {
    h = mws[i](h)
  }
  return h
}