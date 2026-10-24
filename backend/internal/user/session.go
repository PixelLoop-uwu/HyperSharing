package user

import (
  "context"
  "crypto/rand"
  "crypto/sha256"
  "encoding/base64"
  "encoding/hex"
  "encoding/json"
  "errors"
  "fmt"
  "time"

  "github.com/google/uuid"
  "github.com/redis/go-redis/v9"
)

const (
  sessionPrefix      = "session:"
  userSessionsPrefix = "user_sessions:"
)

// sessionValue — то, что лежит в Redis под ключом сессии.
type sessionValue struct {
  UserID    uuid.UUID `json:"user_id"`
  CreatedAt time.Time `json:"created_at"`
}

// SessionStore — непрозрачные сессии в Redis.
type SessionStore struct {
  rdb     *redis.Client
  idleTTL time.Duration // продлевается GETEX при каждом запросе
  maxAge  time.Duration // абсолютный потолок жизни сессии
}

func NewSessionStore(rdb *redis.Client, idleTTL, maxAge time.Duration) *SessionStore {
  return &SessionStore{rdb: rdb, idleTTL: idleTTL, maxAge: maxAge}
}

// Create создаёт сессию и возвращает сырой токен для cookie. Сам токен в Redis не
// хранится — только его SHA-256, чтобы дамп Redis не раскрыл рабочие токены.
func (s *SessionStore) Create(ctx context.Context, userID uuid.UUID) (string, error) {
  token, err := newToken()
  if err != nil {
    return "", fmt.Errorf("create session: %w", err)
  }
  key := sessionKey(token)
  data, err := json.Marshal(sessionValue{UserID: userID, CreatedAt: time.Now()})
  if err != nil {
    return "", fmt.Errorf("create session: %w", err)
  }
  if err := s.rdb.Set(ctx, key, data, s.idleTTL).Err(); err != nil {
    return "", fmt.Errorf("create session: set: %w", err)
  }
  if err := s.rdb.SAdd(ctx, userSessionsKey(userID), key).Err(); err != nil {
    return "", fmt.Errorf("create session: sadd: %w", err)
  }
  // Множество сессий пользователя не должно жить дольше, чем может жить любая
  // сессия (maxAge): протухшие ключи в нём прибираются этим же TTL.
  if err := s.rdb.Expire(ctx, userSessionsKey(userID), s.maxAge).Err(); err != nil {
    return "", fmt.Errorf("create session: expire: %w", err)
  }
  return token, nil
}

// Resolve возвращает user_id по токену, атомарно продлевая TTL (GETEX).
// ErrSessionNotFound, если сессии нет или она старше абсолютного максимума.
func (s *SessionStore) Resolve(ctx context.Context, token string) (uuid.UUID, error) {
  if token == "" {
    return uuid.Nil, ErrSessionNotFound
  }
  key := sessionKey(token)
  data, err := s.rdb.GetEx(ctx, key, s.idleTTL).Bytes()
  if err != nil {
    if errors.Is(err, redis.Nil) {
      return uuid.Nil, ErrSessionNotFound
    }
    return uuid.Nil, fmt.Errorf("resolve session: %w", err)
  }
  var v sessionValue
  if err := json.Unmarshal(data, &v); err != nil {
    return uuid.Nil, fmt.Errorf("resolve session: decode: %w", err)
  }
  if time.Since(v.CreatedAt) > s.maxAge {
    // Сессия пережила абсолютный максимум — активный клиент не должен жить вечно.
    // Протухший ключ в user_sessions приберётся в RevokeAll.
    _ = s.rdb.Del(ctx, key).Err()
    return uuid.Nil, ErrSessionNotFound
  }
  return v.UserID, nil
}

// Revoke удаляет одну сессию по токену. Идемпотентно.
func (s *SessionStore) Revoke(ctx context.Context, token string) error {
  if token == "" {
    return nil
  }
  key := sessionKey(token)
  data, err := s.rdb.Get(ctx, key).Bytes()
  if err != nil {
    if errors.Is(err, redis.Nil) {
      return nil // уже удалена
    }
    return fmt.Errorf("revoke session: get: %w", err)
  }
  var v sessionValue
  _ = json.Unmarshal(data, &v) // даже при повреждённом value ключ всё равно удаляем
  if err := s.rdb.Del(ctx, key).Err(); err != nil {
    return fmt.Errorf("revoke session: del: %w", err)
  }
  if v.UserID != uuid.Nil {
    if err := s.rdb.SRem(ctx, userSessionsKey(v.UserID), key).Err(); err != nil {
      return fmt.Errorf("revoke session: srem: %w", err)
    }
  }
  return nil
}

// RevokeAll удаляет все сессии пользователя. Для «выйти везде» и смены пароля.
func (s *SessionStore) RevokeAll(ctx context.Context, userID uuid.UUID) error {
  setKey := userSessionsKey(userID)
  members, err := s.rdb.SMembers(ctx, setKey).Result()
  if err != nil {
    return fmt.Errorf("revoke all sessions: %w", err)
  }
  if len(members) == 0 {
    return nil
  }
  keys := append([]string{setKey}, members...)
  if err := s.rdb.Del(ctx, keys...).Err(); err != nil {
    return fmt.Errorf("revoke all sessions: %w", err)
  }
  return nil
}

func newToken() (string, error) {
  b := make([]byte, 32)
  if _, err := rand.Read(b); err != nil {
    return "", err
  }
  return base64.RawURLEncoding.EncodeToString(b), nil
}

func sessionKey(token string) string {
  sum := sha256.Sum256([]byte(token))
  return sessionPrefix + hex.EncodeToString(sum[:])
}

func userSessionsKey(userID uuid.UUID) string {
  return userSessionsPrefix + userID.String()
}
