package draft

import (
  "crypto/sha256"
  "encoding/hex"
  "net"

  "github.com/google/uuid"
)

// Owner собирает строку владельца: "user:<uuid>" для залогиненного,
// "ip:<hash>" для анонима. Единственное место, где строится этот формат —
// post и storage сравнивают владельца драфта с этой строкой.
func Owner(userID *uuid.UUID, ip string) string {
  if userID != nil {
    return OwnerUserPrefix + userID.String()
  }
  return OwnerIPPrefix + hashIP(ip)
}

// IPFromAddr достаёт host из "host:port" (или возвращает строку как есть, если
// порта нет).
func IPFromAddr(addr string) string {
  host, _, err := net.SplitHostPort(addr)
  if err != nil {
    return addr
  }
  return host
}

// hashIP — короткий хеш IP. Для IPv6 берём префикс /64, чтобы один пользователь
// не разбивался на несколько счётчиков; сырой IP в ключах не храним.
func hashIP(ip string) string {
  parsed := net.ParseIP(ip)
  if parsed != nil {
    if v6 := parsed.To16(); v6 != nil && parsed.To4() == nil {
      parsed = v6.Mask(net.CIDRMask(64, 128))
    }
    ip = parsed.String()
  }
  sum := sha256.Sum256([]byte(ip))
  return hex.EncodeToString(sum[:8])
}
