package models

import (
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
	"math/rand"
	"time"
)

type ULID struct {
	ID string `gorm:"primaryKey" json:"id"`
}

func (u *ULID) BeforeCreate(_ *gorm.DB) (err error) {
	if u.ID == "" {
		entropy := ulid.Monotonic(rand.Reader, 0)
		u.ID = ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
	}
	return
}
