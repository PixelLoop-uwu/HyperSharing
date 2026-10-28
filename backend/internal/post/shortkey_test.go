package post

import (
  "strings"
  "testing"

  "hypersharing/internal/draft"
)

func TestNewShortKey(t *testing.T) {
  seen := map[string]bool{}
  for i := 0; i < 100; i++ {
    k, err := newShortKey()
    if err != nil {
      t.Fatalf("newShortKey: %v", err)
    }
    if len(k) != shortKeyLen {
      t.Fatalf("len = %d, want %d", len(k), shortKeyLen)
    }
    for _, c := range k {
      if !strings.ContainsRune(shortKeyAlphabet, c) {
        t.Fatalf("char %q outside alphabet", c)
      }
    }
    seen[k] = true
  }
  if len(seen) < 90 {
    t.Fatalf("too many collisions: %d unique of 100", len(seen))
  }
}

func TestOwnerUserID(t *testing.T) {
  if _, err := ownerUserID("ip:abc"); err == nil {
    t.Fatal("non-user owner must error")
  }
  if _, err := ownerUserID(draft.OwnerUserPrefix + "not-a-uuid"); err == nil {
    t.Fatal("malformed user owner must error")
  }
  if _, err := ownerUserID(draft.OwnerUserPrefix + "00000000-0000-0000-0000-000000000000"); err != nil {
    t.Fatal("valid user owner must parse")
  }
}
