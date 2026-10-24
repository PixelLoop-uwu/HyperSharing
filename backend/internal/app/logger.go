package app

import (
  "log/slog"
  "os"
)

func newLogger(debug bool) *slog.Logger {
  if debug {
    return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
  }
  return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}