package google_places

import (
	v1 "api/src/generated/google_maps/v1"
	"api/src/internal/ports"
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"crypto/sha256"

	"log/slog"

	"github.com/valkey-io/valkey-go"
	"golang.org/x/sync/singleflight"
	"google.golang.org/protobuf/proto"
)

type CachedPlacesClient struct {
	inner  ports.PlacesClient // <— the interface from step 1
	kv     valkey.Client
	ttl    time.Duration
	sf     singleflight.Group
	prefix string
}

func NewCachedPlaces(inner ports.PlacesClient, kv valkey.Client, ttl time.Duration) *CachedPlacesClient {
	return &CachedPlacesClient{inner: inner, kv: kv, ttl: ttl, prefix: "gplaces:v1:"}
}

var _ ports.PlacesClient = (*CachedPlacesClient)(nil)

func (c *CachedPlacesClient) get(ctx context.Context, key string, dst proto.Message) (bool, error) {
	r := c.kv.Do(ctx, c.kv.B().Get().Key(key).Build())
	if err := r.Error(); err != nil {
		// Check if this is a "key not found" error rather than a real error
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

func (c *CachedPlacesClient) set(ctx context.Context, key string, msg proto.Message) {
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

func (c *CachedPlacesClient) cachedFetch(ctx context.Context, key string, dst proto.Message, fetchFn func() (proto.Message, error)) (proto.Message, error) {
	if ok, _ := c.get(ctx, key, dst); ok {
		return dst, nil
	}

	result, err, _ := c.sf.Do(key, func() (any, error) {
		if ok, _ := c.get(ctx, key, dst); ok {
			return dst, nil
		}
		fresh, err := fetchFn()
		if err != nil {
			return nil, err
		}
		if fresh == nil {
			return nil, fmt.Errorf("fetchFn returned nil result")
		}
		c.set(ctx, key, fresh)
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

func (c *CachedPlacesClient) cachedFetchPlace(ctx context.Context, key string, fetchFn func() (*v1.Place, error)) (*v1.Place, error) {
	result, err := c.cachedFetch(ctx, key, &v1.Place{}, func() (proto.Message, error) {
		return fetchFn()
	})
	if err != nil {
		return nil, err
	}
	return result.(*v1.Place), nil
}

func (c *CachedPlacesClient) cachedFetchSearchResponse(ctx context.Context, key string, fetchFn func() (*v1.SearchTextResponse, error)) (*v1.SearchTextResponse, error) {
	result, err := c.cachedFetch(ctx, key, &v1.SearchTextResponse{}, func() (proto.Message, error) {
		return fetchFn()
	})
	if err != nil {
		return nil, err
	}
	return result.(*v1.SearchTextResponse), nil
}

func hashKey(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte{0})
		h.Write([]byte(p))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (c *CachedPlacesClient) buildKey(operation string, params ...string) string {
	return c.prefix + operation + ":" + hashKey(params...)
}

func normalizeValue(v any) string {
	switch val := v.(type) {
	case []string:
		sorted := make([]string, len(val))
		copy(sorted, val)
		sort.Strings(sorted)
		return "[" + strings.Join(sorted, ",") + "]"
	case []v1.PriceLevel:
		strs := make([]string, len(val))
		for i, level := range val {
			strs[i] = strconv.Itoa(int(level))
		}
		sort.Strings(strs)
		return "[" + strings.Join(strs, ",") + "]"
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case bool:
		return strconv.FormatBool(val)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int64:
		return strconv.FormatInt(val, 10)
	case int:
		return strconv.Itoa(val)
	default:
		return fmt.Sprint(val)
	}
}

func (c *CachedPlacesClient) keyForRequest(operation string, params map[string]any) string {
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
		parts = append(parts, k+"="+normalizeValue(v))
	}
	return c.buildKey(operation, parts...)
}

func (c *CachedPlacesClient) keyGetPlace(req *v1.GetPlaceRequest) string {
	return c.keyForRequest("get", map[string]any{
		"name":   req.Name,
		"lang":   req.LanguageCode,
		"region": req.RegionCode,
		"fields": req.RequestedFields,
	})
}

func (c *CachedPlacesClient) keyGetRestaurantDetails(req *v1.GetRestaurantDetailsRequest) string {
	return c.keyForRequest("get_restaurant", map[string]any{
		"name":   req.Name,
		"lang":   req.LanguageCode,
		"region": req.RegionCode,
	})
}

func (c *CachedPlacesClient) keySearchText(req *v1.SearchTextRequest) string {
	return c.keyForRequest("search_text", map[string]any{
		"q":      req.TextQuery,
		"lang":   req.LanguageCode,
		"region": req.RegionCode,
		"rank":   req.RankPreference.String(),
		"type":   req.IncludedType,
		"open":   req.OpenNow,
		"min":    req.MinRating,
		"max":    req.MaxResultCount,
		"prices": req.PriceLevels,
		"strict": req.StrictTypeFiltering,
		"pure":   req.IncludePureServiceAreaBusinesses,
		"fields": req.RequestedFields,
	})
}

func (c *CachedPlacesClient) keySearchRestaurants(req *v1.SearchRestaurantsRequest) string {
	return c.keyForRequest("search_restaurants", map[string]any{
		"q":      req.TextQuery,
		"lang":   req.LanguageCode,
		"region": req.RegionCode,
	})
}

func (c *CachedPlacesClient) GetRestaurantDetails(ctx context.Context, req *v1.GetRestaurantDetailsRequest) (*v1.Place, error) {
	key := c.keyGetRestaurantDetails(req)
	return c.cachedFetchPlace(ctx, key, func() (*v1.Place, error) {
		return c.inner.GetRestaurantDetails(ctx, req)
	})
}

func (c *CachedPlacesClient) GetPlace(ctx context.Context, req *v1.GetPlaceRequest) (*v1.Place, error) {
	key := c.keyGetPlace(req)
	return c.cachedFetchPlace(ctx, key, func() (*v1.Place, error) {
		return c.inner.GetPlace(ctx, req)
	})
}

func (c *CachedPlacesClient) SearchText(ctx context.Context, req *v1.SearchTextRequest) (*v1.SearchTextResponse, error) {
	key := c.keySearchText(req)
	return c.cachedFetchSearchResponse(ctx, key, func() (*v1.SearchTextResponse, error) {
		return c.inner.SearchText(ctx, req)
	})
}

func (c *CachedPlacesClient) SearchRestaurants(ctx context.Context, req *v1.SearchRestaurantsRequest) (*v1.SearchTextResponse, error) {
	key := c.keySearchRestaurants(req)
	return c.cachedFetchSearchResponse(ctx, key, func() (*v1.SearchTextResponse, error) {
		return c.inner.SearchRestaurants(ctx, req)
	})
}
