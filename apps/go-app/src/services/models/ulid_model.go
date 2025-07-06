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
		u.ID = ulid.MustNew(ulid.Timestamp(time.Now()), rand.New(rand.NewSource(time.Now().UnixNano()))).String()
	}
	return
}
