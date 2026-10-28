// Package worker — фоновая очистка: брошенные tus-загрузки и сиротские объекты
// в S3. Запускается отдельным бинарём cmd/worker. Источник правды для «живых»
// объектов — БД (files.storage_key); возраст объектов отсекает ещё не
// опубликованные файлы активных драфтов.
package worker

import (
  "context"
  "log/slog"
  "time"

  "hypersharing/internal/platform/s3x"
  "hypersharing/internal/post"
  "hypersharing/internal/user"
)

type Config struct {
  Interval       time.Duration
  AbandonedAge   time.Duration // старше этого — брошенная загрузка
  OrphanAge      time.Duration // старше этого — сиротский объект (должно быть > DraftTTL)
  ObjectPrefix   string        // files/ — финальные объекты
  MetadataPrefix string        // tusmeta/ — .info/.part
  AvatarPrefix   string        // avatars/ — аватары
}

type Worker struct {
  repo   *post.Repository
  users  *user.Repository
  bucket *s3x.Bucket
  cfg    Config
}

func New(repo *post.Repository, users *user.Repository, bucket *s3x.Bucket, cfg Config) *Worker {
  return &Worker{repo: repo, users: users, bucket: bucket, cfg: cfg}
}

// Run прогоняет джобы сразу и затем по тикеру, пока ctx не отменят.
func (w *Worker) Run(ctx context.Context) error {
  w.runAll(ctx)
  ticker := time.NewTicker(w.cfg.Interval)
  defer ticker.Stop()
  for {
    select {
    case <-ctx.Done():
      return nil
    case <-ticker.C:
      w.runAll(ctx)
    }
  }
}

func (w *Worker) runAll(ctx context.Context) {
  jobs := []struct {
    name string
    fn   func(context.Context) error
  }{
    {"abandoned_uploads", w.cleanupAbandonedUploads},
    {"orphan_objects", w.cleanupOrphanObjects},
    {"orphan_avatars", w.cleanupOrphanAvatars},
  }
  for _, j := range jobs {
    if err := j.fn(ctx); err != nil {
      slog.Error("worker job failed", "job", j.name, "err", err)
    }
  }
}

// cleanupAbandonedUploads отменяет брошенные multipart-загрузки и удаляет
// старые .info/.part. Активная загрузка освежает .info на каждом PATCH, поэтому
// возраст > AbandonedAge означает «клиент бросил».
func (w *Worker) cleanupAbandonedUploads(ctx context.Context) error {
  cutoff := time.Now().Add(-w.cfg.AbandonedAge)

  uploads, err := w.bucket.ListIncompleteUploads(ctx, w.cfg.ObjectPrefix)
  if err != nil {
    return err
  }
  for _, u := range uploads {
    if u.Initiated.After(cutoff) {
      continue
    }
    if err := w.bucket.AbortUpload(ctx, u.Key); err != nil {
      slog.Error("abort upload", "key", u.Key, "err", err)
    }
  }

  metas, err := w.bucket.List(ctx, w.cfg.MetadataPrefix)
  if err != nil {
    return err
  }
  for _, m := range metas {
    if m.LastModified.After(cutoff) {
      continue
    }
    if err := w.bucket.Remove(ctx, m.Key); err != nil {
      slog.Error("remove meta object", "key", m.Key, "err", err)
    }
  }
  return nil
}

// cleanupOrphanObjects удаляет объекты files/, которых нет ни в одном
// опубликованном посте и которые старше OrphanAge (значит, не принадлежат
// активному драфту — драфты живут короче).
func (w *Worker) cleanupOrphanObjects(ctx context.Context) error {
  keys, err := w.repo.AllStorageKeys(ctx)
  if err != nil {
    return err
  }
  live := make(map[string]struct{}, len(keys))
  for _, k := range keys {
    live[k] = struct{}{}
  }

  objs, err := w.bucket.List(ctx, w.cfg.ObjectPrefix)
  if err != nil {
    return err
  }
  for _, key := range orphanKeys(objs, live, time.Now().Add(-w.cfg.OrphanAge)) {
    if err := w.bucket.Remove(ctx, key); err != nil {
      slog.Error("remove orphan", "key", key, "err", err)
    }
  }
  return nil
}

// cleanupOrphanAvatars удаляет аватары, чей пользователь снял аватар или удалён
// (нет записи в users.avatar_version). Возраст защищает от гонки между S3 Put и
// записью версии в БД.
func (w *Worker) cleanupOrphanAvatars(ctx context.Context) error {
  refs, err := w.users.AvatarUserIDs(ctx)
  if err != nil {
    return err
  }
  live := make(map[string]struct{}, len(refs))
  for _, ref := range refs {
    live[user.AvatarKey(w.cfg.AvatarPrefix, ref.ID, ref.Version)] = struct{}{}
  }

  objs, err := w.bucket.List(ctx, w.cfg.AvatarPrefix)
  if err != nil {
    return err
  }
  cutoff := time.Now().Add(-w.cfg.OrphanAge)
  for _, o := range objs {
    if _, ok := live[o.Key]; ok {
      continue
    }
    if o.LastModified.After(cutoff) {
      continue
    }
    if err := w.bucket.Remove(ctx, o.Key); err != nil {
      slog.Error("remove orphan avatar", "key", o.Key, "err", err)
    }
  }
  return nil
}

// orphanKeys возвращает ключи объектов старше cutoff, которых нет в live.
func orphanKeys(objs []s3x.ObjectInfo, live map[string]struct{}, cutoff time.Time) []string {
  var out []string
  for _, o := range objs {
    if _, ok := live[o.Key]; ok {
      continue
    }
    if o.LastModified.After(cutoff) {
      continue
    }
    out = append(out, o.Key)
  }
  return out
}
