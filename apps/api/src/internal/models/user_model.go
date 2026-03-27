package models

import (
	userpb "api/src/generated/users/v1"
	"time"

	"gorm.io/gorm"
)

type User struct {
	UUIDv7
	GoogleId           *string   `gorm:"uniqueIndex"`
	Email              *string   `gorm:"uniqueIndex"`
	Username           *string   `gorm:"uniqueIndex"`
	Name               string    `gorm:"not null"`
	IsDarkModeEnabled  bool      `gorm:"default:false"`
	DefaultRegion      string    `gorm:"default:''"`
	DefaultLanguage    string    `gorm:"default:''"`
	CreatedAt          time.Time `gorm:"autoCreateTime"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if err = u.UUIDv7.BeforeCreate(tx); err != nil {
		return err
	}
	return nil
}

func (u *User) ToProto() *userpb.UserProto {
	return &userpb.UserProto{
		Id:                u.ID,
		GoogleId:          derefString(u.GoogleId),
		Email:             derefString(u.Email),
		Username:          derefString(u.Username),
		Name:              u.Name,
		IsDarkModeEnabled: u.IsDarkModeEnabled,
		DefaultRegion:     u.DefaultRegion,
		DefaultLanguage:   u.DefaultLanguage,
		CreatedAt:         u.CreatedAt.Unix(),
		UpdatedAt:         u.UpdatedAt.Unix(),
	}
}

// StringPtr returns a pointer to s, or nil if s is empty.
func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
