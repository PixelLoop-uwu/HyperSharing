//go:build integration

package s3x

import (
  "bytes"
  "context"
  "errors"
  "testing"

  "hypersharing/internal/testutil"
)

func newBucket(t *testing.T) *Bucket {
  t.Helper()
  s3 := testutil.MinIO(t)
  c, err := New(Config{Endpoint: s3.Endpoint, AccessKey: s3.AccessKey, SecretKey: s3.SecretKey, UseSSL: false, Region: "us-east-1"})
  if err != nil {
    t.Fatalf("s3x new: %v", err)
  }
  if err := c.EnsureBucket(context.Background(), "test-bucket"); err != nil {
    t.Fatalf("ensure bucket: %v", err)
  }
  return c.Bucket("test-bucket")
}

func TestBucketPutGetRemove(t *testing.T) {
  b := newBucket(t)
  ctx := context.Background()

  data := []byte("hello world")
  info, err := b.Put(ctx, "files/abc", bytes.NewReader(data), int64(len(data)), "text/plain")
  if err != nil {
    t.Fatalf("put: %v", err)
  }
  if info.Size != int64(len(data)) {
    t.Fatalf("size: %d", info.Size)
  }

  obj, got, err := b.Get(ctx, "files/abc")
  if err != nil {
    t.Fatalf("get: %v", err)
  }
  buf := new(bytes.Buffer)
  if _, err := buf.ReadFrom(obj); err != nil {
    t.Fatalf("read: %v", err)
  }
  obj.Close()
  if buf.String() != string(data) {
    t.Fatalf("content: %q", buf.String())
  }
  if got.Size != int64(len(data)) {
    t.Fatalf("size: %d", got.Size)
  }

  if err := b.Remove(ctx, "files/abc"); err != nil {
    t.Fatalf("remove: %v", err)
  }
  if _, _, err := b.Get(ctx, "files/abc"); !errors.Is(err, ErrNotFound) {
    t.Fatalf("want not found after remove, got %v", err)
  }
}

func TestBucketList(t *testing.T) {
  b := newBucket(t)
  ctx := context.Background()

  for _, k := range []string{"files/a", "files/b", "avatars/x"} {
    if _, err := b.Put(ctx, k, bytes.NewReader([]byte("x")), 1, "text/plain"); err != nil {
      t.Fatalf("put %s: %v", k, err)
    }
  }
  objs, err := b.List(ctx, "files/")
  if err != nil {
    t.Fatalf("list: %v", err)
  }
  if len(objs) != 2 {
    t.Fatalf("want 2 objects under files/, got %d", len(objs))
  }
}
