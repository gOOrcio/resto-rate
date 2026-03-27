package services

import (
	tagsv1 "api/src/generated/tags/v1"
	"api/src/generated/tags/v1/v1connect"
	"api/src/internal/cache"
	"api/src/internal/models"
	"context"
	"time"

	"errors"

	"connectrpc.com/connect"
	"github.com/valkey-io/valkey-go"
	"gorm.io/gorm"
)

const tagsCacheKey = "tags:all"
const tagsCacheTTL = time.Hour

type TagsService struct {
	v1connect.UnimplementedTagsServiceHandler
	DB     *gorm.DB
	Valkey valkey.Client
}

func NewTagsService(db *gorm.DB, kv valkey.Client) *TagsService {
	return &TagsService{DB: db, Valkey: kv}
}

func (s *TagsService) ListTags(
	ctx context.Context,
	_ *connect.Request[tagsv1.ListTagsRequest],
) (*connect.Response[tagsv1.ListTagsResponse], error) {
	// Try cache first (skip if Valkey is nil, e.g. in tests)
	if s.Valkey != nil {
		pc := cache.NewProtoCache(s.Valkey, tagsCacheTTL, "")
		cached := &tagsv1.ListTagsResponse{}
		if ok, _ := pc.Get(ctx, tagsCacheKey, cached); ok {
			return connect.NewResponse(cached), nil
		}
	}

	if s.DB == nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("database not initialized"))
	}

	var tags []models.Tag
	if err := s.DB.WithContext(ctx).Order("category, label").Find(&tags).Error; err != nil {
		return nil, err
	}

	protos := make([]*tagsv1.TagProto, len(tags))
	for i, t := range tags {
		protos[i] = t.ToProto()
	}
	resp := &tagsv1.ListTagsResponse{Tags: protos}

	// Populate cache
	if s.Valkey != nil {
		pc := cache.NewProtoCache(s.Valkey, tagsCacheTTL, "")
		pc.Set(ctx, tagsCacheKey, resp)
	}

	return connect.NewResponse(resp), nil
}
