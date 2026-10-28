package post

import (
  "errors"
  "testing"
)

func testLimits() Limits {
  return Limits{
    MaxBytes:      1000,
    MaxFiles:      5,
    MaxSections:   5,
    MaxDrafts:     3,
    TitleMaxLen:   10,
    SectionMaxLen: 10,
    StorageQuota:  10000,
  }
}

func TestValidateDraft(t *testing.T) {
  lim := testLimits()

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
    {"no sections", withSections(ok, 0), errValidation()},
    {"title too long", withTitle(ok, string(make([]byte, 11))), errValidation()},
    {"too many sections", withSections(ok, 6), errValidation()},
    {"too many files", withFiles(ok, 6), errValidation()},
    {"file too large", withFileSize(ok, 1001), errValidation()},
    {"post too large", withTotal(ok, 1001), errValidation()},
    {"section too long", withBody(ok, string(make([]byte, 11))), errValidation()},
    {"empty name", withName(ok, ""), errValidation()},
  }

  for _, c := range cases {
    t.Run(c.name, func(t *testing.T) {
      _, err := validateDraft(lim, c.in)
      if c.want == nil && err != nil {
        t.Fatalf("unexpected error: %v", err)
      }
      if c.want != nil && !isValidationErr(err) {
        t.Fatalf("want ValidationError, got %v", err)
      }
    })
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
  // два файла по total-1 и 1 — сумма total, каждый ≤ limit
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
