package user

import (
  "crypto/rand"
  "crypto/subtle"
  "encoding/base64"
  "fmt"
  "runtime"
  "strings"

  "golang.org/x/crypto/argon2"
)

// argon2Sem ограничивает число одновременных дорогих вычислений argon2 (каждое
// держит Memory KiB). Семафор размером с число ядер не даёт сотне параллельных
// логинов/регистраций выесть всю память процесса.
var argon2Sem = make(chan struct{}, runtime.GOMAXPROCS(0))

// Argon2Params — параметры argon2id. Живут в PHC-строке вместе с хешем,
// поэтому их можно поднимать со временем, не ломая старые пароли.
type Argon2Params struct {
  Memory      uint32 // KiB
  Iterations  uint32
  Parallelism uint8
  SaltLen     uint32
  KeyLen      uint32
}

// hashPassword возвращает PHC-строку вида
// $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>.
func hashPassword(password string, p Argon2Params) (string, error) {
  argon2Sem <- struct{}{}
  defer func() { <-argon2Sem }()

  salt := make([]byte, p.SaltLen)
  if _, err := rand.Read(salt); err != nil {
    return "", fmt.Errorf("hash password: read salt: %w", err)
  }
  hash := argon2.IDKey([]byte(password), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLen)
  return encodePHC(salt, hash, p), nil
}

// verifyPassword проверяет пароль против PHC-строки. Параметры берутся из самой
// строки. Сравнение — в постоянном времени.
func verifyPassword(password, encoded string) (bool, error) {
  p, salt, expected, err := decodePHC(encoded)
  if err != nil {
    return false, err
  }
  argon2Sem <- struct{}{}
  defer func() { <-argon2Sem }()
  actual := argon2.IDKey([]byte(password), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLen)
  return subtle.ConstantTimeCompare(actual, expected) == 1, nil
}

func encodePHC(salt, hash []byte, p Argon2Params) string {
  return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
    argon2.Version,
    p.Memory, p.Iterations, p.Parallelism,
    base64.RawStdEncoding.EncodeToString(salt),
    base64.RawStdEncoding.EncodeToString(hash),
  )
}

func decodePHC(encoded string) (Argon2Params, []byte, []byte, error) {
  // ["", "argon2id", "v=19", "m=65536,t=3,p=4", salt, hash]
  parts := strings.Split(encoded, "$")
  if len(parts) != 6 || parts[1] != "argon2id" {
    return Argon2Params{}, nil, nil, fmt.Errorf("decode phc: invalid format")
  }
  var version int
  if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
    return Argon2Params{}, nil, nil, fmt.Errorf("decode phc: unsupported version")
  }
  var p Argon2Params
  if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &p.Parallelism); err != nil {
    return Argon2Params{}, nil, nil, fmt.Errorf("decode phc: bad params: %w", err)
  }
  salt, err := base64.RawStdEncoding.DecodeString(parts[4])
  if err != nil {
    return Argon2Params{}, nil, nil, fmt.Errorf("decode phc: bad salt: %w", err)
  }
  hash, err := base64.RawStdEncoding.DecodeString(parts[5])
  if err != nil {
    return Argon2Params{}, nil, nil, fmt.Errorf("decode phc: bad hash: %w", err)
  }
  p.SaltLen = uint32(len(salt))
  p.KeyLen = uint32(len(hash))
  return p, salt, hash, nil
}
