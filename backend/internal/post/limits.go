package post

// Limits — лимиты постов и драфтов залогиненного пользователя. Заполняется из
// env-конфига в app.go.
type Limits struct {
  MaxBytes      int64
  MaxFiles      int
  MaxSections   int
  MaxDrafts     int
  TitleMaxLen   int
  SectionMaxLen int
  StorageQuota  int64
}
