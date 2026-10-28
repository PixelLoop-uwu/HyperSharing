package draft

import (
  "net"

  "github.com/google/uuid"
)

// Owner собирает строку владельца: "user:<uuid>". Единственное место, где
// строится этот формат — post и storage сравнивают владельца драфта с этой строкой.
func Owner(userID uuid.UUID) string {
  return OwnerUserPrefix + userID.String()
}

// IPFromAddr достаёт host из "host:port" (или возвращает строку как есть, если
// порта нет). Нужен для rate limit по IP клиента.
func IPFromAddr(addr string) string {
  host, _, err := net.SplitHostPort(addr)
  if err != nil {
    return addr
  }
  return host
}
