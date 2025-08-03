package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UUIDv7 struct {
	ID string `gorm:"primaryKey" json:"id"`
}

func (u *UUIDv7) BeforeCreate(_ *gorm.DB) (err error) {
	if u.ID == "" {
		newUUID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		u.ID = newUUID.String()
	}
	return
}
