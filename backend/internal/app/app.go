package app

import (
  "context"
  "fmt"
  "log/slog"
  "net/http"
  "os"
  "os/signal"
  "syscall"
  "time"

  "github.com/go-chi/chi/v5"

  "hypersharing/internal/config"
  "hypersharing/internal/draft"
  "hypersharing/internal/httpx"
  "hypersharing/internal/platform/postgres"
  "hypersharing/internal/platform/ratelimit"
  "hypersharing/internal/platform/redisx"
  "hypersharing/internal/platform/s3x"
  "hypersharing/internal/post"
  "hypersharing/internal/storage"
  "hypersharing/internal/user"
)

const shutdownTimeout = 10 * time.Second
const apiPrefix = "/api/v1"

func Run(ctx context.Context) error {
  ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
  defer stop()

  cfg, err := config.Load()
  if err != nil {
    return fmt.Errorf("load config: %w", err)
  }
  slog.SetDefault(newLogger(cfg.Debug))

  pool, err := postgres.New(ctx, cfg.DatabaseURL)
  if err != nil {
    return err
  }
  defer pool.Close()

  rdb, err := redisx.New(ctx, cfg.RedisURL)
  if err != nil {
    return err
  }
  defer rdb.Close()

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

  // user feature
  userRepo := user.NewRepository(pool)
  sessions := user.NewSessionStore(rdb, cfg.Session.TTL, cfg.Session.MaxAge)
  userSvc := user.NewService(userRepo, sessions, user.Argon2Params{
    Memory:      cfg.Argon2.Memory,
    Iterations:  cfg.Argon2.Iterations,
    Parallelism: cfg.Argon2.Parallelism,
    SaltLen:     cfg.Argon2.SaltLen,
    KeyLen:      cfg.Argon2.KeyLen,
  })
  limiter := ratelimit.New(rdb)
  userHandler := user.NewHandler(userSvc, cfg.Session.CookieSecure, cfg.Session.MaxAge, limiter, cfg.RateLimit.LoginEmail, cfg.RateLimit.Window)

  avatarService := user.NewAvatarService(userRepo, s3c.Bucket(cfg.S3.Bucket), limiter, user.AvatarConfig{
    MaxBytes: cfg.Avatar.MaxBytes,
    MaxDim:   cfg.Avatar.MaxDim,
    Size:     cfg.Avatar.Size,
    Quality:  cfg.Avatar.Quality,
    Rate:     cfg.Avatar.Rate,
    Window:   cfg.Avatar.Window,
    Prefix:   cfg.Avatar.Prefix,
  })

  // Общее хранилище драфтов (владелец схемы — пакет draft). Его используют и
  // post (создание/удаление), и storage (статусы слотов при загрузке).
  draftStore := draft.NewStore(rdb, cfg.Posts.DraftTTL)

  // storage feature
  storageSvc, err := storage.New(storage.Config{
    S3EndpointURL:       cfg.S3.Endpoint,
    S3AccessKeyID:       cfg.S3.AccessKeyID,
    S3SecretAccessKey:   cfg.S3.SecretAccessKey,
    S3Region:            cfg.S3.Region,
    S3PathStyle:         cfg.S3.PathStyle,
    Bucket:              cfg.S3.Bucket,
    ObjectPrefix:        cfg.Uploads.ObjectPrefix,
    MetadataPrefix:      cfg.Uploads.MetadataPrefix,
    DailyBytes:        cfg.Quota.DailyBytes,
    FileNotFoundLimit: cfg.RateLimit.FileNotFound,
    DownloadLimit:     cfg.RateLimit.Download,
    RateLimitWindow:   cfg.RateLimit.Window,
    TusBasePath:       apiPrefix + "/uploads/",
  }, rdb, pool, s3c.Bucket(cfg.S3.Bucket), draftStore, limiter)
  if err != nil {
    return err
  }

  // post feature
  postSvc := post.New(pool, rdb, storageSvc.Quota(), draftStore, post.Limits{
    MaxBytes:      cfg.Posts.MaxBytes,
    MaxFiles:      cfg.Posts.MaxFiles,
    MaxSections:   cfg.Posts.MaxSections,
    MaxDrafts:     cfg.Posts.MaxDrafts,
    TitleMaxLen:   cfg.Posts.TitleMaxLen,
    SectionMaxLen: cfg.Posts.SectionMaxLen,
    StorageQuota:  cfg.Posts.StorageQuota,
  }, apiPrefix)
  postHandler := post.NewHandler(postSvc, limiter, cfg.RateLimit.Draft, cfg.RateLimit.Window)

  api := chi.NewRouter()
  api.Use(httpx.OriginCheck(cfg.AllowedOrigins))
  api.Route("/auth", func(r chi.Router) {
    r.With(limiter.Middleware(rateKey("rl:register:"), cfg.RateLimit.Register, cfg.RateLimit.Window)).Post("/register", userHandler.Register)
    r.With(limiter.Middleware(rateKey("rl:login:"), cfg.RateLimit.Login, cfg.RateLimit.Window)).Post("/login", userHandler.Login)
    r.Post("/logout", userHandler.Logout)
    r.With(user.RequireAuth(userSvc)).Get("/me", userHandler.Me)
  })

  api.Route("/me/avatar", func(r chi.Router) {
    r.With(user.RequireAuth(userSvc)).Put("/", avatarService.UploadHandler)
    r.With(user.RequireAuth(userSvc)).Delete("/", avatarService.DeleteHandler)
  })
  api.Get("/avatars/{userID}", avatarService.Serve)

  api.Route("/drafts", func(r chi.Router) {
    r.With(user.RequireAuth(userSvc)).Post("/", postHandler.CreateDraft)
    r.With(user.RequireAuth(userSvc)).Delete("/{id}", postHandler.CancelDraft)
    r.With(user.RequireAuth(userSvc)).Post("/{id}/publish", postHandler.Publish)
  })
  api.Get("/posts/{shortKey}", postHandler.GetByShortKey)
  api.With(user.RequireAuth(userSvc)).Delete("/posts/{id}", postHandler.Delete)
  api.With(user.RequireAuth(userSvc)).Get("/me/posts", postHandler.ListMine)
  api.With(user.RequireAuth(userSvc)).Get("/me/quota", postHandler.Quota)

  // выдача файлов по короткому ключу
  api.Get("/f/{shortKey}", storageSvc.ServeFile())
  api.Head("/f/{shortKey}", storageSvc.ServeFile())

  mux := http.NewServeMux()
  mux.HandleFunc("GET /health", health(map[string]check{
    "postgres": pool.Ping,
    "redis":    func(ctx context.Context) error { return rdb.Ping(ctx).Err() },
    "s3":       func(ctx context.Context) error { return s3c.Ping(ctx, cfg.S3.Bucket) },
  }))
  mux.Handle(apiPrefix+"/", http.StripPrefix(apiPrefix, api))

  srv := &http.Server{
    Addr:              cfg.Addr,
    Handler:           httpx.Chain(mux, httpx.RequestID, httpx.AccessLog, httpx.Recover),
    ReadHeaderTimeout: 5 * time.Second,
    ReadTimeout:       cfg.ReadTimeout,
    WriteTimeout:      cfg.WriteTimeout,
    IdleTimeout:       60 * time.Second,
  }

  // tus-загрузки на отдельном сервере: у них долгие PATCH, поэтому общий
  // ReadTimeout/WriteTimeout к ним не применяется (иначе легитимная загрузка
  // обрывалась бы). Порядок тот же, что был внутри api-роутера:
  // OriginCheck → RequireAuth → tus.
  tus := http.StripPrefix(apiPrefix+"/uploads", storageSvc.Tus())
  tus = httpx.OriginCheck(cfg.AllowedOrigins)(user.RequireAuth(userSvc)(tus))
  tusSrv := &http.Server{
    Addr:              cfg.TusAddr,
    Handler:           httpx.Chain(tus, httpx.RequestID, httpx.AccessLog, httpx.Recover),
    ReadHeaderTimeout: 5 * time.Second,
    IdleTimeout:       60 * time.Second,
  }

  srvErr := make(chan error, 2)
  go func() { srvErr <- srv.ListenAndServe() }()
  go func() { srvErr <- tusSrv.ListenAndServe() }()
  slog.Info("listening", "addr", cfg.Addr, "tus_addr", cfg.TusAddr)

  select {
  case err := <-srvErr:
    return fmt.Errorf("http server: %w", err) // не смог стартовать, например порт занят
  case <-ctx.Done():
  }

  slog.Info("shutting down")
  sctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
  defer cancel()
  if err := srv.Shutdown(sctx); err != nil {
    return fmt.Errorf("shutdown: %w", err)
  }
  if err := tusSrv.Shutdown(sctx); err != nil {
    return fmt.Errorf("shutdown tus: %w", err)
  }
  return nil
}

// rateKey строит ключ rate limit по IP клиента (RemoteAddr — доверенный прокси
// не поддерживаем, см. «Защита от злоупотреблений» в Claude.md).
func rateKey(prefix string) func(*http.Request) string {
  return func(r *http.Request) string { return prefix + draft.IPFromAddr(r.RemoteAddr) }
}
