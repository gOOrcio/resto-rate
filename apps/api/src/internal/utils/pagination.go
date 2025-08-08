package utils

import (
	"fmt"

	"gorm.io/gorm"
)

// PaginationResult contains the result of pagination with metadata
type PaginationResult struct {
	TotalCount    int64
	Page          int32
	PageSize      int32
	NextPageToken string
}

// Paginate performs pagination on a database query and returns results with metadata
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

// CalculateNextPageToken calculates the next page token for pagination
// This function prevents integer overflow by casting all components to int64
func CalculateNextPageToken(page, pageSize int32, currentItemsCount int, totalCount int64) string {
	// Cast all components to int64 to prevent integer overflow
	currentItemsTotal := int64(page-1)*int64(pageSize) + int64(currentItemsCount)
	if currentItemsTotal < totalCount {
		return fmt.Sprintf("%d", page+1)
	}
	return ""
}
