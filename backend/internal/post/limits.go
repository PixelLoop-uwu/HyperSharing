package post

import (
  "strings"
  "time"

  "hypersharing/internal/draft"
)

// Limits — лимиты постов, отдельно для анонимов и залогиненных. Заполняется
// из env-конфига в app.go.
type Limits struct {
  AnonMaxBytes    int64
  UserMaxBytes    int64
  AnonMaxFiles    int
  UserMaxFiles    int
  AnonMaxSections int
  UserMaxSections int
  AnonMaxDrafts   int
  UserMaxDrafts   int
  TitleMaxLen     int
  SectionMaxLen   int
  AnonTTL         time.Duration
  DraftTTL        time.Duration
  UserStorageQuota int64
}

func (l Limits) isUser(owner string) bool {
  return strings.HasPrefix(owner, draft.OwnerUserPrefix)
}

func (l Limits) maxBytes(owner string) int64 {
  if l.isUser(owner) {
    return l.UserMaxBytes
  }
  return l.AnonMaxBytes
}

func (l Limits) maxFiles(owner string) int {
  if l.isUser(owner) {
    return l.UserMaxFiles
  }
  return l.AnonMaxFiles
}

func (l Limits) maxSections(owner string) int {
  if l.isUser(owner) {
    return l.UserMaxSections
  }
  return l.AnonMaxSections
}

func (l Limits) maxDrafts(owner string) int {
  if l.isUser(owner) {
    return l.UserMaxDrafts
  }
  return l.AnonMaxDrafts
}
