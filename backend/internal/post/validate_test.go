package post

import (
  "errors"
  "testing"
  "time"

  "hypersharing/internal/draft"
)

func testLimits() Limits {
  return Limits{
    AnonMaxBytes:     100,
    UserMaxBytes:     1000,
    AnonMaxFiles:     2,
    UserMaxFiles:     5,
    AnonMaxSections:  2,
    UserMaxSections:  5,
    AnonMaxDrafts:    2,
    UserMaxDrafts:    3,
    TitleMaxLen:      10,
    SectionMaxLen:    10,
    AnonTTL:          time.Hour,
    DraftTTL:         time.Hour,
    UserStorageQuota: 10000,
  }
}

func TestValidateDraft(t *testing.T) {
  lim := testLimits()
  anon := draft.OwnerIPPrefix + "x"

  ok := CreateDraftInput{
    Title:    "title",
    Sections: []SectionInput{{Body: "text", Files: []FileInput{{Name: "a.txt", Mime: "text/plain", Size: 10}}}},
  }

  cases := []struct {
    name string
    in   CreateDraftInput
    want error
  }{
    {"ok", ok, nil},
    {"title too long", withTitle(ok, string(make([]byte, 11))), errValidation()},
    {"too many sections", withSections(ok, 3), errValidation()},
    {"too many files", withFiles(ok, 3), errValidation()},
    {"file too large", withFileSize(ok, 101), errValidation()},
    {"post too large", withTotal(ok, 101), errValidation()},
    {"section too long", withBody(ok, string(make([]byte, 11))), errValidation()},
    {"empty name", withName(ok, ""), errValidation()},
  }

  for _, c := range cases {
    t.Run(c.name, func(t *testing.T) {
      _, err := validateDraft(lim, anon, c.in)
      if c.want == nil && err != nil {
        t.Fatalf("unexpected error: %v", err)
      }
      if c.want != nil && !isValidationErr(err) {
        t.Fatalf("want ValidationError, got %v", err)
      }
    })
  }
}

func TestValidateDraftUserLimitsHigher(t *testing.T) {
  lim := testLimits()
  user := draft.OwnerUserPrefix + "00000000-0000-0000-0000-000000000000"

  // 3 файла: для анонима превышает (2), для юзера — нет (5).
  in := withFiles(CreateDraftInput{Sections: []SectionInput{{Files: []FileInput{
    {Name: "a", Mime: "x", Size: 1},
    {Name: "b", Mime: "x", Size: 1},
    {Name: "c", Mime: "x", Size: 1},
  }}}}, 3)

  if _, err := validateDraft(lim, draft.OwnerIPPrefix+"x", in); !isValidationErr(err) {
    t.Fatalf("anon should reject 3 files, got %v", err)
  }
  if _, err := validateDraft(lim, user, in); err != nil {
    t.Fatalf("user should accept 3 files, got %v", err)
  }
}

// --- helpers to tweak a single field ---

func errValidation() error { return &ValidationError{Msg: "x"} }

func isValidationErr(err error) bool {
  var vErr *ValidationError
  return errors.As(err, &vErr)
}

func withTitle(in CreateDraftInput, title string) CreateDraftInput {
  in.Title = title
  return in
}

func withSections(in CreateDraftInput, n int) CreateDraftInput {
  in.Sections = make([]SectionInput, n)
  for i := range in.Sections {
    in.Sections[i] = SectionInput{Body: "x"}
  }
  return in
}

func withFiles(in CreateDraftInput, n int) CreateDraftInput {
  files := make([]FileInput, n)
  for i := range files {
    files[i] = FileInput{Name: "a", Mime: "x", Size: 1}
  }
  in.Sections = []SectionInput{{Files: files}}
  return in
}

func withFileSize(in CreateDraftInput, size int64) CreateDraftInput {
  in.Sections = []SectionInput{{Files: []FileInput{{Name: "a", Mime: "x", Size: size}}}}
  return in
}

func withTotal(in CreateDraftInput, total int64) CreateDraftInput {
  // два файла по 51 — сумма 102 > 100, но каждый ≤ 100
  in.Sections = []SectionInput{{Files: []FileInput{
    {Name: "a", Mime: "x", Size: total - 1},
    {Name: "b", Mime: "x", Size: 1},
  }}}
  return in
}

func withBody(in CreateDraftInput, body string) CreateDraftInput {
  in.Sections = []SectionInput{{Body: body}}
  return in
}

func withName(in CreateDraftInput, name string) CreateDraftInput {
  in.Sections = []SectionInput{{Files: []FileInput{{Name: name, Mime: "x", Size: 1}}}}
  return in
}
