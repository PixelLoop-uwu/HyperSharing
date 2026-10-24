package post

import (
  "testing"
  "time"

  "github.com/google/uuid"
)

func TestCursorRoundTrip(t *testing.T) {
  original := time.Date(2026, 10, 24, 12, 30, 45, 123456000, time.UTC)
  id := uuid.New()

  gotT, gotID, err := decodeCursor(encodeCursor(original, id))
  if err != nil {
    t.Fatalf("decode: %v", err)
  }
  if !gotT.Equal(original) {
    t.Fatalf("time: got %v, want %v", gotT, original)
  }
  if gotID != id {
    t.Fatalf("id: got %v, want %v", gotID, id)
  }
}

func TestDecodeCursorRejectsGarbage(t *testing.T) {
  for _, s := range []string{"", "not-base64!", "!!!", "AAAA"} {
    if _, _, err := decodeCursor(s); err == nil {
      t.Fatalf("expected error for %q", s)
    }
  }
}
