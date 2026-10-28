package main

import (
  "context"
  "log/slog"
  "os"
  "os/signal"
  "syscall"

  "hypersharing/internal/config"
  "hypersharing/internal/platform/postgres"
  "hypersharing/internal/platform/s3x"
  "hypersharing/internal/post"
  "hypersharing/internal/user"
  "hypersharing/internal/worker"
)

func main() {
  if err := run(); err != nil {
    slog.Error("fatal", "err", err)
    os.Exit(1)
  }
}

func run() error {
  ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
  defer stop()

  cfg, err := config.Load()
  if err != nil {
    return err
  }
  slog.SetDefault(newLogger(cfg.Debug))

  pool, err := postgres.New(ctx, cfg.DatabaseURL)
  if err != nil {
    return err
  }
  defer pool.Close()

  s3Host, s3UseSSL, err := cfg.S3.HostPort()
  if err != nil {
    return err
  }
  s3c, err := s3x.New(s3x.Config{
    Endpoint:  s3Host,
    AccessKey: cfg.S3.AccessKeyID,
    SecretKey: cfg.S3.SecretAccessKey,
    UseSSL:    s3UseSSL,
    Region:    cfg.S3.Region,
  })
  if err != nil {
    return err
  }

  w := worker.New(post.NewRepository(pool), user.NewRepository(pool), s3c.Bucket(cfg.S3.Bucket), worker.Config{
    Interval:       cfg.Worker.Interval,
    AbandonedAge:   cfg.Worker.AbandonedUploadAge,
    OrphanAge:      cfg.Worker.OrphanObjectAge,
    ObjectPrefix:   cfg.Uploads.ObjectPrefix,
    MetadataPrefix: cfg.Uploads.MetadataPrefix,
    AvatarPrefix:   cfg.Avatar.Prefix,
  })

  slog.Info("worker started", "interval", cfg.Worker.Interval)
  return w.Run(ctx)
}

func newLogger(debug bool) *slog.Logger {
  if debug {
    return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
  }
  return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
