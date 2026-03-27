package models

import (
	tagsv1 "api/src/generated/tags/v1"
	"time"

	"gorm.io/gorm"
)

type Tag struct {
	UUIDv7
	Slug      string    `gorm:"uniqueIndex;not null"`
	Label     string    `gorm:"not null"`
	Category  string    `gorm:"not null;index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (t *Tag) BeforeCreate(tx *gorm.DB) (err error) {
	return t.UUIDv7.BeforeCreate(tx)
}

func (t *Tag) ToProto() *tagsv1.TagProto {
	return &tagsv1.TagProto{
		Id:       t.ID,
		Slug:     t.Slug,
		Label:    t.Label,
		Category: t.Category,
	}
}
