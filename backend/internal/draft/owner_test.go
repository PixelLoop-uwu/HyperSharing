package draft

import (
  "testing"

  "github.com/google/uuid"
)

func TestOwnerUserVsAnon(t *testing.T) {
  id := uuid.New()
  if Owner(&id, "1.2.3.4") != OwnerUserPrefix+id.String() {
    t.Fatalf("user owner must ignore IP")
  }
  if Owner(nil, "1.2.3.4") != OwnerIPPrefix+hashIP("1.2.3.4") {
    t.Fatalf("anon owner must hash IP")
  }
}

func TestOwnerCollapsesIPv6To64(t *testing.T) {
  // один /64 → одинаковый владелец
  a := Owner(nil, "2001:db8:1:2::1")
  b := Owner(nil, "2001:db8:1:2::ffff")
  if a != b {
    t.Fatalf("same /64 must produce equal owner: %q vs %q", a, b)
  }
}

func TestIPFromAddr(t *testing.T) {
  if got := IPFromAddr("1.2.3.4:1234"); got != "1.2.3.4" {
    t.Fatalf("want 1.2.3.4, got %q", got)
  }
  if got := IPFromAddr("2001:db8::1"); got != "2001:db8::1" {
    t.Fatalf("addr without port must pass through, got %q", got)
  }
}
