package user

import (
  "errors"
  "time"

  "github.com/google/uuid"
)

// User — пользователь, как его видит API и контекст запроса.
// password_hash сюда не попадает: он живёт только в UserAuth.
type User struct {
  ID            uuid.UUID `json:"id"`
  Username      string    `json:"username"`
  Email         string    `json:"email"`
  EmailVerified bool      `json:"email_verified"`
  IsAdmin       bool      `json:"is_admin"`
  CreatedAt     time.Time `json:"created_at"`
  AvatarVersion *string   `json:"avatar_version"`

}

// UserAuth — пользователь вместе с хешем пароля. Только для проверки при входе.
type UserAuth struct {
  User
  PasswordHash string
}

// RegisterInput — тело POST /auth/register.
type RegisterInput struct {
  Username string `json:"username"`
  Email    string `json:"email"`
  Password string `json:"password"`
}

// LoginInput — тело POST /auth/login (вход по email).
type LoginInput struct {
  Email    string `json:"email"`
  Password string `json:"password"`
}

// ValidationError — невалидные входные данные. Handler отвечает на неё 400.
type ValidationError struct {
  Msg string
}

func (e *ValidationError) Error() string { return e.Msg }

// Доменные ошибки. В HTTP-статусы их переводит только handler.
var (
  ErrNotFound           = errors.New("user: not found")
  ErrEmailTaken         = errors.New("user: email already taken")
  ErrUsernameTaken      = errors.New("user: username already taken")
  ErrInvalidCredentials = errors.New("user: invalid email or password")
  ErrSessionNotFound    = errors.New("user: session not found")
)
