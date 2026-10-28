package draft

import (
  "testing"

  "github.com/google/uuid"
)

func TestOwner(t *testing.T) {
  id := uuid.New()
  if Owner(id) != OwnerUserPrefix+id.String() {
    t.Fatalf("owner = %q, want %q", Owner(id), OwnerUserPrefix+id.String())
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
