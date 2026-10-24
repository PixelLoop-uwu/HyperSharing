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
  if _, ok := ownerUserID(draft.OwnerIPPrefix + "abc"); ok {
    t.Fatal("ip owner must not parse as user")
  }
  if _, ok := ownerUserID(draft.OwnerUserPrefix + "not-a-uuid"); ok {
    t.Fatal("malformed user owner must not parse")
  }
  if _, ok := ownerUserID(draft.OwnerUserPrefix + "00000000-0000-0000-0000-000000000000"); !ok {
    t.Fatal("valid user owner must parse")
  }
}
