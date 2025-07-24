package models

import (
	userpb "api/src/generated/users/v1"
	"gorm.io/gorm"
	"time"
)

type User struct {
	ULID
	GoogleId  string    `gorm:"uniqueIndex"`
	Email     string    `gorm:"uniqueIndex"`
	Username  string    `gorm:"uniqueIndex"`
	Name      string    `gorm:"not null"`
	IsAdmin   bool      `gorm:"default:false"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if err = u.ULID.BeforeCreate(tx); err != nil {
		return err
	}
	return nil
}

func (u *User) ToProto() *userpb.UserProto {
	return &userpb.UserProto{
		Id:        u.ID,
		GoogleId:  u.GoogleId,
		Email:     u.Email,
		Name:      u.Name,
		IsAdmin:   u.IsAdmin,
		CreatedAt: u.CreatedAt.Unix(),
		UpdatedAt: u.UpdatedAt.Unix(),
	}
}
