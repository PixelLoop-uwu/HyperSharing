package worker

import (
  "testing"
  "time"

  "hypersharing/internal/platform/s3x"
)

func TestOrphanKeys(t *testing.T) {
  now := time.Now()
  cutoff := now.Add(-48 * time.Hour)
  old := now.Add(-72 * time.Hour)
  young := now.Add(-1 * time.Hour)

  objs := []s3x.ObjectInfo{
    {Key: "files/live-old", LastModified: old},     // в БД → не трогаем
    {Key: "files/orphan-old", LastModified: old},   // нет в БД, старый → удалить
    {Key: "files/orphan-young", LastModified: young}, // нет в БД, свежий → не трогаем
  }
  live := map[string]struct{}{"files/live-old": {}}

  got := orphanKeys(objs, live, cutoff)
  want := []string{"files/orphan-old"}
  if len(got) != 1 || got[0] != want[0] {
    t.Fatalf("got %v, want %v", got, want)
  }
}
