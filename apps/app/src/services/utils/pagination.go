package utils

import (
	"gorm.io/gorm"
)

func Paginate(db *gorm.DB, model interface{}, page, pageSize int32) (int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var totalCount int64
	if err := db.Model(model).Count(&totalCount).Error; err != nil {
		return 0, err
	}

	if err := db.Limit(int(pageSize)).Offset(int(offset)).Find(model).Error; err != nil {
		return 0, err
	}

	return totalCount, nil
}
