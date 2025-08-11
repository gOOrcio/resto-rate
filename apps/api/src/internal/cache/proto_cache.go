package cache

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/valkey-io/valkey-go"
	"golang.org/x/sync/singleflight"
	"google.golang.org/protobuf/proto"
)

// ProtoCache provides generic caching for protocol buffer messages
type ProtoCache struct {
	kv     valkey.Client
	ttl    time.Duration
	sf     singleflight.Group
	prefix string
}

// NewProtoCache creates a new ProtoCache instance
func NewProtoCache(kv valkey.Client, ttl time.Duration, prefix string) *ProtoCache {
	return &ProtoCache{kv: kv, ttl: ttl, prefix: prefix}
}

// Get retrieves a proto message from the cache
func (c *ProtoCache) Get(ctx context.Context, key string, dst proto.Message) (bool, error) {
	r := c.kv.Do(ctx, c.kv.B().Get().Key(key).Build())
	if err := r.Error(); err != nil {
		if strings.Contains(err.Error(), "valkey nil message") {
			slog.Debug("Cache key not found", slog.String("key", key))
			return false, nil
		}
		slog.Debug("Cache get error", slog.String("key", key), slog.Any("error", err))
		return false, nil
	}
	raw, err := r.AsBytes()
	if err != nil {
		slog.Debug("Cache value decode error", slog.String("key", key), slog.Any("error", err))
		return false, nil
	}
	if len(raw) == 0 {
		slog.Debug("Cache key not found", slog.String("key", key))
		return false, nil
	}
	if err := proto.Unmarshal(raw, dst); err != nil {
		slog.Debug("Cache proto unmarshal error", slog.String("key", key), slog.Any("error", err))
		return false, nil
	}
	slog.Debug("Cache hit", slog.String("key", key))
	return true, nil
}

// Set stores a proto message in the cache
func (c *ProtoCache) Set(ctx context.Context, key string, msg proto.Message) {
	if msg == nil {
		slog.Debug("Cache set skipped - nil message", slog.String("key", key))
		return
	}

	raw, err := proto.MarshalOptions{Deterministic: true}.Marshal(msg)
	if err != nil {
		slog.Debug("Cache marshal error", slog.String("key", key), slog.Any("error", err))
		return
	}

	if c.ttl > 0 {
		if err := c.kv.Do(ctx, c.kv.B().Set().
			Key(key).Value(valkey.BinaryString(raw)).
			Ex(c.ttl).
			Build()).Error(); err != nil {
			slog.Debug("Cache set error", slog.String("key", key), slog.Any("error", err))
			return
		}
	} else {
		if err := c.kv.Do(ctx, c.kv.B().Set().
			Key(key).Value(valkey.BinaryString(raw)).
			Build()).Error(); err != nil {
			slog.Debug("Cache set error", slog.String("key", key), slog.Any("error", err))
			return
		}
	}

	slog.Debug("Cache set", slog.String("key", key))
}

// CachedFetch implements the pattern of fetch-with-cache with single flight deduplication
func (c *ProtoCache) CachedFetch(ctx context.Context, key string, dst proto.Message, fetchFn func() (proto.Message, error)) (proto.Message, error) {
	if ok, _ := c.Get(ctx, key, dst); ok {
		return dst, nil
	}

	result, err, _ := c.sf.Do(key, func() (any, error) {
		if ok, _ := c.Get(ctx, key, dst); ok {
			return dst, nil
		}
		fresh, err := fetchFn()
		if err != nil {
			return nil, err
		}
		if fresh == nil {
			return nil, fmt.Errorf("fetchFn returned nil result")
		}
		c.Set(ctx, key, fresh)
		return fresh, nil
	})

	if err != nil {
		return nil, err
	}

	typed, ok := result.(proto.Message)
	if !ok || typed == nil {
		return nil, fmt.Errorf("singleflight returned unexpected type %T", result)
	}
	return typed, nil
}

// BuildKey builds a cache key with the given prefix and operation
func (c *ProtoCache) BuildKey(operation string, parts ...string) string {
	return c.prefix + operation + ":" + HashKey(parts...)
}

// KeyForRequest builds a cache key from a map of request parameters
func (c *ProtoCache) KeyForRequest(operation string, params map[string]any) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(params))
	for _, k := range keys {
		v := params[k]
		if v == nil || (fmt.Sprint(v) == "" || fmt.Sprint(v) == "[]") {
			continue
		}
		parts = append(parts, k+"="+NormalizeValue(v))
	}
	return c.BuildKey(operation, parts...)
}
