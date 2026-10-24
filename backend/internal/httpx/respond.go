package httpx

import (
  "encoding/json"
  "errors"
  "net/http"
)

const maxBodyBytes = 1 << 20 // 1 MiB

type errorBody struct {
  Error errorDetail `json:"error"`
}

type errorDetail struct {
  Code    string `json:"code"`
  Message string `json:"message"`
}

// JSON writes v as a JSON response with the given status.
func JSON(w http.ResponseWriter, status int, v any) {
  buf, err := json.Marshal(v)
  if err != nil {
    Logger(nil).Error("encode response", "err", err)
    Error(w, http.StatusInternalServerError, "internal", "internal error")
    return
  }
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)
  w.Write(buf)
}

// Error writes a client-facing error in the common format.
func Error(w http.ResponseWriter, status int, code, message string) {
  JSON(w, status, errorBody{errorDetail{Code: code, Message: message}})
}

// Internal logs the real cause and sends the client a generic 500.
func Internal(w http.ResponseWriter, r *http.Request, err error) {
  Logger(r.Context()).Error("internal error", "err", err)
  Error(w, http.StatusInternalServerError, "internal", "internal error")
}

// DecodeJSON reads the request body into T. On failure it writes the error
// response itself and returns ok=false, so the handler just returns.
func DecodeJSON[T any](w http.ResponseWriter, r *http.Request) (v T, ok bool) {
  r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
  if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
    var tooBig *http.MaxBytesError
    if errors.As(err, &tooBig) {
      Error(w, http.StatusRequestEntityTooLarge, "body_too_large", "request body too large")
    } else {
      Error(w, http.StatusBadRequest, "bad_json", "invalid JSON body")
    }
    return v, false
  }
  return v, true
}