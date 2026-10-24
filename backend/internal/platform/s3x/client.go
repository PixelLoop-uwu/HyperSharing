// Package s3x — тонкая обёртка над minio-go.
//
// Остальной код проекта не должен импортировать minio напрямую: если захотим
// сменить клиента (например, на aws-sdk-go-v2), правки ограничатся этим пакетом.
package s3x

import (
  "context"
  "fmt"

  "github.com/minio/minio-go/v7"
  "github.com/minio/minio-go/v7/pkg/credentials"
)

// Config — параметры подключения к S3-совместимому хранилищу (MinIO).
type Config struct {
  Endpoint  string // host:port без схемы, например "localhost:9000"
  AccessKey string
  SecretKey string
  UseSSL    bool
  Region    string // для MinIO можно оставить пустым
}

// Client держит соединение с хранилищем. Безопасен для конкурентного использования.
type Client struct {
  mc     *minio.Client
  region string
}

// New создаёт клиента. Сеть здесь не трогается (minio-go подключается лениво),
// поэтому сразу после New вызывай Ping или EnsureBucket — fail fast при старте.
func New(cfg Config) (*Client, error) {
  mc, err := minio.New(cfg.Endpoint, &minio.Options{
    Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
    Secure: cfg.UseSSL,
    Region: cfg.Region,
  })
  if err != nil {
    return nil, fmt.Errorf("s3x: new client: %w", err)
  }
  return &Client{mc: mc, region: cfg.Region}, nil
}

// Ping проверяет, что хранилище доступно и бакет существует. Для /readyz.
func (c *Client) Ping(ctx context.Context, bucket string) error {
  ok, err := c.mc.BucketExists(ctx, bucket)
  if err != nil {
    return fmt.Errorf("s3x: ping: %w", err)
  }
  if !ok {
    return fmt.Errorf("s3x: bucket %q does not exist", bucket)
  }
  return nil
}

// EnsureBucket создаёт бакет, если его ещё нет. Идемпотентно.
// В docker-compose бакеты уже создаёт minio-init, так что это страховка для dev;
// в проде бакеты лучше создавать вне приложения.
func (c *Client) EnsureBucket(ctx context.Context, bucket string) error {
  ok, err := c.mc.BucketExists(ctx, bucket)
  if err != nil {
    return fmt.Errorf("s3x: check bucket %q: %w", bucket, err)
  }
  if ok {
    return nil
  }

  err = c.mc.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: c.region})
  if err != nil {
    // Гонка: другой экземпляр успел создать бакет между проверкой и созданием.
    if minio.ToErrorResponse(err).Code == "BucketAlreadyOwnedByYou" {
      return nil
    }
    return fmt.Errorf("s3x: make bucket %q: %w", bucket, err)
  }
  return nil
}

// Bucket возвращает хэндл на конкретный бакет. Дёшево, ничего не проверяет.
// Фичи получают свой бакет через конструктор: storage — uploads, users — avatars.
func (c *Client) Bucket(name string) *Bucket {
  return &Bucket{mc: c.mc, name: name}
}