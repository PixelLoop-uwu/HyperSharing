package user

import (
  "context"
  "errors"
  "time"

  "github.com/google/uuid"
  "github.com/jackc/pgx/v5"
  "github.com/jackc/pgx/v5/pgconn"
  "github.com/jackc/pgx/v5/pgtype"

  "hypersharing/internal/user/db"
)

// Repository — доступ к данным пользователей поверх сгенерированного sqlc.
type Repository struct {
  q *db.Queries
}

func NewRepository(dbtx db.DBTX) *Repository {
  return &Repository{q: db.New(dbtx)}
}

func (r *Repository) CreateUser(ctx context.Context, username, email, passwordHash string) (User, error) {
  row, err := r.q.CreateUser(ctx, db.CreateUserParams{
    Username:     username,
    Email:        email,
    PasswordHash: passwordHash,
  })
  if err != nil {
    return User{}, mapPgErr(err)
  }
  return toUser(row.ID, row.Username, row.Email, row.EmailVerified, row.IsAdmin, row.CreatedAt, pgtype.Text{}), nil
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (User, error) {
  row, err := r.q.GetUserByID(ctx, id)
  if err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
      return User{}, ErrNotFound
    }
    return User{}, err
  }
  return toUser(row.ID, row.Username, row.Email, row.EmailVerified, row.IsAdmin, row.CreatedAt, row.AvatarVersion), nil
}

func (r *Repository) GetUserAuthByEmail(ctx context.Context, email string) (UserAuth, error) {
  row, err := r.q.GetUserAuthByEmail(ctx, email)
  if err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
      return UserAuth{}, ErrNotFound
    }
    return UserAuth{}, err
  }
  return UserAuth{
    User:         toUser(row.ID, row.Username, row.Email, row.EmailVerified, row.IsAdmin, row.CreatedAt, row.AvatarVersion),
    PasswordHash: row.PasswordHash,
  }, nil
}

func toUser(id uuid.UUID, username, email string, verified, admin bool, createdAt time.Time, avatarVersion pgtype.Text) User {
  var avatar *string
  if avatarVersion.Valid {
    avatar = &avatarVersion.String
  }
  return User{
    ID:            id,
    Username:      username,
    Email:         email,
    EmailVerified: verified,
    IsAdmin:       admin,
    CreatedAt:     createdAt,
    AvatarVersion: avatar,
  }
}

// mapPgErr превращает нарушение уникальности в доменную ошибку, чтобы service не
// разбирал коды Postgres сам. Остальные ошибки проходят как есть.
func mapPgErr(err error) error {
  var pgErr *pgconn.PgError
  if errors.As(err, &pgErr) && pgErr.Code == "23505" {
    switch pgErr.ConstraintName {
    case "users_email_uniq":
      return ErrEmailTaken
    case "users_username_uniq":
      return ErrUsernameTaken
    }
  }
  return err
}

// --- Аватары ---

func (r *Repository) SetAvatarVersion(ctx context.Context, id uuid.UUID, version string) error {
  return r.q.SetAvatarVersion(ctx, db.SetAvatarVersionParams{ID: id, Version: pgtype.Text{String: version, Valid: true}})
}

func (r *Repository) ClearAvatarVersion(ctx context.Context, id uuid.UUID) error {
  return r.q.ClearAvatarVersion(ctx, id)
}

// GetAvatarVersion возвращает версию аватара. ErrNotFound, если пользователя
// нет или аватара нет (avatar_version NULL).
func (r *Repository) GetAvatarVersion(ctx context.Context, id uuid.UUID) (string, error) {
  v, err := r.q.GetAvatarVersion(ctx, id)
  if err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
      return "", ErrNotFound
    }
    return "", err
  }
  if !v.Valid {
    return "", ErrNotFound
  }
  return v.String, nil
}

// AvatarRef — id пользователя и текущая версия его аватара (ключ объекта в S3).
type AvatarRef struct {
  ID      uuid.UUID
  Version string
}

// AvatarUserIDs возвращает пользователей с аватаром и их текущие версии. Нужен
// воркеру для GC сиротских аватаров.
func (r *Repository) AvatarUserIDs(ctx context.Context) ([]AvatarRef, error) {
  rows, err := r.q.AvatarUserIDs(ctx)
  if err != nil {
    return nil, err
  }
  refs := make([]AvatarRef, len(rows))
  for i, row := range rows {
    refs[i] = AvatarRef{ID: row.ID, Version: row.AvatarVersion.String}
  }
  return refs, nil
}
