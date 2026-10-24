package post

import (
  "crypto/rand"
  "math/big"
)

// short_key — 8 символов в публичном URL. Алфавит a-z A-Z 0-9 (62 символа),
// без знаков препинания, чтобы ключ был безобиден в URL и путях.
const shortKeyAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

const shortKeyLen = 8

// newShortKey генерирует криптостойкий ключ без модульного смещения
// (rejection sampling через rand.Int). Коллизия на UNIQUE обрабатывается
// провалом публикации — вероятность ~1/5.7e10 на ключ, ретрай не нужен.
func newShortKey() (string, error) {
  b := make([]byte, shortKeyLen)
  n := big.NewInt(int64(len(shortKeyAlphabet)))
  for i := range b {
    v, err := rand.Int(rand.Reader, n)
    if err != nil {
      return "", err
    }
    b[i] = shortKeyAlphabet[v.Int64()]
  }
  return string(b), nil
}
