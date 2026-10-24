package user

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
  // Слабые параметры, чтобы тест был быстрым.
  p := Argon2Params{Memory: 1024, Iterations: 1, Parallelism: 1, SaltLen: 16, KeyLen: 32}

  encoded, err := hashPassword("correct horse battery staple", p)
  if err != nil {
    t.Fatalf("hash: %v", err)
  }

  if ok, err := verifyPassword("correct horse battery staple", encoded); err != nil || !ok {
    t.Fatalf("verify correct: ok=%v err=%v", ok, err)
  }
  if ok, _ := verifyPassword("wrong", encoded); ok {
    t.Fatal("verify wrong: expected false")
  }
}

func TestVerifyPasswordUsesEmbeddedParams(t *testing.T) {
  // Параметры читаются из PHC-строки, а не из внешнего конфига.
  p := Argon2Params{Memory: 2048, Iterations: 2, Parallelism: 2, SaltLen: 16, KeyLen: 32}
  encoded, err := hashPassword("password", p)
  if err != nil {
    t.Fatalf("hash: %v", err)
  }
  if ok, err := verifyPassword("password", encoded); err != nil || !ok {
    t.Fatalf("verify: ok=%v err=%v", ok, err)
  }
}

func TestVerifyPasswordRejectsMalformed(t *testing.T) {
  for _, s := range []string{"", "not-a-phc", "$argon2id$v=19$m=1,t=1,p=1", "$argon2id$v=18$m=1,t=1,p=1$x$y"} {
    if _, err := verifyPassword("x", s); err == nil {
      t.Errorf("expected error for %q", s)
    }
  }
}
