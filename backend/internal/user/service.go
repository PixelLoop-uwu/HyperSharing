package user

import (
  "context"
  "errors"
  "regexp"
  "strings"

  "github.com/google/uuid"
)

// userRepository — то, что service нужно от БД. Объявлен на стороне потребителя.
type userRepository interface {
  CreateUser(ctx context.Context, username, email, passwordHash string) (User, error)
  GetUserByID(ctx context.Context, id uuid.UUID) (User, error)
  GetUserAuthByEmail(ctx context.Context, email string) (UserAuth, error)
}

type Service struct {
  repo      userRepository
  sessions  *SessionStore
  argon     Argon2Params
  dummyHash string // выравнивает время ответа login при несуществующем email
}

func NewService(repo userRepository, sessions *SessionStore, p Argon2Params) *Service {
  s := &Service{repo: repo, sessions: sessions, argon: p}
  // Один раз считаем «пустой» хеш: login с несуществующим email гоняет argon2 на
  // нём, чтобы не отличать «нет такого email» от «неверный пароль» по таймингу.
  if h, err := hashPassword("dummy-password-for-timing", p); err == nil {
    s.dummyHash = h
  }
  return s
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (User, string, error) {
  if err := validateRegister(in); err != nil {
    return User{}, "", err
  }
  hash, err := hashPassword(in.Password, s.argon)
  if err != nil {
    return User{}, "", err
  }
  u, err := s.repo.CreateUser(ctx, strings.TrimSpace(in.Username), normalizeEmail(in.Email), hash)
  if err != nil {
    return User{}, "", err
  }
  token, err := s.sessions.Create(ctx, u.ID)
  if err != nil {
    return User{}, "", err
  }
  return u, token, nil
}

func (s *Service) Login(ctx context.Context, in LoginInput) (User, string, error) {
  if err := validateLogin(in); err != nil {
    return User{}, "", err
  }
  email := normalizeEmail(in.Email)
  ua, err := s.repo.GetUserAuthByEmail(ctx, email)
  if err != nil {
    if errors.Is(err, ErrNotFound) {
      // Пользователя нет — всё равно тратим то же время на argon2 (см. dummyHash).
      _, _ = verifyPassword(in.Password, s.dummyHash)
      return User{}, "", ErrInvalidCredentials
    }
    return User{}, "", err
  }
  ok, err := verifyPassword(in.Password, ua.PasswordHash)
  if err != nil {
    return User{}, "", err
  }
  if !ok {
    return User{}, "", ErrInvalidCredentials
  }
  token, err := s.sessions.Create(ctx, ua.ID)
  if err != nil {
    return User{}, "", err
  }
  return ua.User, token, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
  return s.sessions.Revoke(ctx, token)
}

// ResolveSession возвращает пользователя по токену сессии. (nil, nil) — если
// токена нет или сессия недействительна; ошибка — только при реальном сбое.
func (s *Service) ResolveSession(ctx context.Context, token string) (*User, error) {
  if token == "" {
    return nil, nil
  }
  userID, err := s.sessions.Resolve(ctx, token)
  if err != nil {
    if errors.Is(err, ErrSessionNotFound) {
      return nil, nil
    }
    return nil, err
  }
  u, err := s.repo.GetUserByID(ctx, userID)
  if err != nil {
    if errors.Is(err, ErrNotFound) {
      return nil, nil // пользователь удалён — сессия мертва
    }
    return nil, err
  }
  return &u, nil
}

var (
  usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
  emailRe    = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
)

func validateRegister(in RegisterInput) error {
  switch {
  case len(in.Username) < 3 || len(in.Username) > 32:
    return &ValidationError{"username must be 3..32 characters"}
  case !usernameRe.MatchString(in.Username):
    return &ValidationError{"username may contain only letters, digits, '-' and '_'"}
  case len(in.Email) > 254 || !emailRe.MatchString(in.Email):
    return &ValidationError{"invalid email"}
  case len(in.Password) < 8 || len(in.Password) > 128:
    return &ValidationError{"password must be 8..128 characters"}
  }
  return nil
}

func validateLogin(in LoginInput) error {
  if len(in.Email) > 254 || !emailRe.MatchString(in.Email) {
    return &ValidationError{"invalid email"}
  }
  if len(in.Password) > 128 {
    return &ValidationError{"password too long"}
  }
  return nil
}

func normalizeEmail(email string) string {
  return strings.ToLower(strings.TrimSpace(email))
}
