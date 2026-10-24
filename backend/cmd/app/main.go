package main

import (
  "context"
  "log/slog"
  "os"

  "hypersharing/internal/app"
)

func main() {
  if err := app.Run(context.Background()); err != nil {
    slog.Error("fatal", "err", err)
    os.Exit(1)
  }
}