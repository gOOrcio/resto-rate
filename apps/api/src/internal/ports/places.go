package ports

import (
	v1 "api/src/generated/google_maps/v1"
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"crypto/sha256"

	"github.com/valkey-io/valkey-go"
	"golang.org/x/sync/singleflight"
	"google.golang.org/protobuf/proto"
)

type CachedPlaces struct {
	inner  PlacesClient
	kv     valkey.Client
	ttl    time.Duration
	sf     singleflight.Group
	prefix string
}

type PlacesClient interface {
	SearchText(ctx context.Context, req *v1.SearchTextRequest) (*v1.SearchTextResponse, error)
	SearchRestaurants(ctx context.Context, req *v1.SearchRestaurantsRequest) (*v1.SearchTextResponse, error)
	GetPlace(ctx context.Context, req *v1.GetPlaceRequest) (*v1.Place, error)
	GetRestaurantDetails(ctx context.Context, req *v1.GetRestaurantDetailsRequest) (*v1.Place, error)
}

func NewCachedPlaces(inner PlacesClient, kv valkey.Client, ttl time.Duration, logger *log.Logger) *CachedPlaces {
	return &CachedPlaces{
		inner:  inner,
		kv:     kv,
		ttl:    ttl,
		prefix: "gplaces:v1:",
	}
}


func (c *CachedPlaces) get(ctx context.Context, key string, dst proto.Message) (bool, error) {
	r := c.kv.Do(ctx, c.kv.B().Get().Key(key).Build())
	if err := r.Error(); err != nil {
		log.Printf("[DEBUG] Cache get error - key: %s, error: %v", key, err)
		return false, nil
	}
	raw, err := r.AsBytes()
	if err != nil || len(raw) == 0 {
		if err != nil {
			log.Printf("[DEBUG] Cache value decode error - key: %s, error: %v", key, err)
		}
		return false, nil
	}
	if err := proto.Unmarshal(raw, dst); err != nil {
		log.Printf("[DEBUG] Cache proto unmarshal error - key: %s, error: %v", key, err)
		return false, nil
	}
	log.Printf("[DEBUG] Cache hit - key: %s", key)
	return true, nil
}

func (c *CachedPlaces) set(ctx context.Context, key string, msg proto.Message) {
	raw, err := proto.MarshalOptions{Deterministic: true}.Marshal(msg)
	if err != nil {
		log.Printf("[DEBUG] Cache marshal error - key: %s, error: %v", key, err)
		return
	}
	
	if c.ttl > 0 {
		if err := c.kv.Do(ctx, c.kv.B().Set().
			Key(key).Value(string(raw)).
			Ex(c.ttl).
			Build()).Error(); err != nil {
			log.Printf("[DEBUG] Cache set error - key: %s, error: %v", key, err)
			return
		}
	} else {
		if err := c.kv.Do(ctx, c.kv.B().Set().
			Key(key).Value(string(raw)).
			Build()).Error(); err != nil {
			log.Printf("[DEBUG] Cache set error - key: %s, error: %v", key, err)
			return
		}
	}
	
	log.Printf("[DEBUG] Cache set - key: %s", key)
}

func (c *CachedPlaces) cachedFetchPlace(ctx context.Context, key string, fetchFn func() (*v1.Place, error)) (*v1.Place, error) {
	var out v1.Place
	if ok, _ := c.get(ctx, key, &out); ok {
		return &out, nil
	}

	ch := c.sf.DoChan(key, func() (any, error) {
		if ok, _ := c.get(ctx, key, &out); ok {
			return &out, nil
		}
		fresh, err := fetchFn()
		if err != nil {
			return nil, err
		}
		c.set(ctx, key, fresh)
		return fresh, nil
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-ch:
		if result.Err != nil {
			return nil, result.Err
		}
		p, ok := result.Val.(*v1.Place)
		if !ok || p == nil {
			return nil, fmt.Errorf("singleflight returned unexpected type %T", result.Val)
		}
		return p, nil
	}
}

func (c *CachedPlaces) cachedFetchSearchResponse(ctx context.Context, key string, fetchFn func() (*v1.SearchTextResponse, error)) (*v1.SearchTextResponse, error) {
	var out v1.SearchTextResponse
	if ok, _ := c.get(ctx, key, &out); ok {
		return &out, nil
	}

	ch := c.sf.DoChan(key, func() (any, error) {
		if ok, _ := c.get(ctx, key, &out); ok {
			return &out, nil
		}
		fresh, err := fetchFn()
		if err != nil {
			return nil, err
		}
		c.set(ctx, key, fresh)
		return fresh, nil
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-ch:
		if result.Err != nil {
			return nil, result.Err
		}
		resp, ok := result.Val.(*v1.SearchTextResponse)
		if !ok || resp == nil {
			return nil, fmt.Errorf("singleflight returned unexpected type %T", result.Val)
		}
		return resp, nil
	}
}

func hashKey(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte{0})
		h.Write([]byte(p))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (c *CachedPlaces) buildKey(operation string, params ...string) string {
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

func (c *CachedPlaces) keyForRequest(operation string, params map[string]any) string {
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

func (c *CachedPlaces) keyGetPlace(req *v1.GetPlaceRequest) string {
	return c.keyForRequest("get", map[string]any{
		"name":   req.Name,
		"lang":   req.LanguageCode,
		"region": req.RegionCode,
		"fields": req.RequestedFields,
	})
}

func (c *CachedPlaces) keyGetRestaurantDetails(req *v1.GetRestaurantDetailsRequest) string {
	return c.keyForRequest("get_restaurant", map[string]any{
		"name":   req.Name,
		"lang":   req.LanguageCode,
		"region": req.RegionCode,
	})
}

func (c *CachedPlaces) keySearchText(req *v1.SearchTextRequest) string {
	return c.keyForRequest("search_text", map[string]any{
		"q":        req.TextQuery,
		"lang":     req.LanguageCode,
		"region":   req.RegionCode,
		"rank":     req.RankPreference.String(),
		"type":     req.IncludedType,
		"open":     req.OpenNow,
		"min":      req.MinRating,
		"max":      req.MaxResultCount,
		"prices":   req.PriceLevels,
		"strict":   req.StrictTypeFiltering,
		"pure":     req.IncludePureServiceAreaBusinesses,
		"fields":   req.RequestedFields,
	})
}

func (c *CachedPlaces) keySearchRestaurants(req *v1.SearchRestaurantsRequest) string {
	return c.keyForRequest("search_restaurants", map[string]any{
		"q":      req.TextQuery,
		"lang":   req.LanguageCode,
		"region": req.RegionCode,
	})
}

func (c *CachedPlaces) GetRestaurantDetails(ctx context.Context, req *v1.GetRestaurantDetailsRequest) (*v1.Place, error) {
	key := c.keyGetRestaurantDetails(req)
	return c.cachedFetchPlace(ctx, key, func() (*v1.Place, error) {
		return c.inner.GetRestaurantDetails(ctx, req)
	})
}

func (c *CachedPlaces) GetPlace(ctx context.Context, req *v1.GetPlaceRequest) (*v1.Place, error) {
	key := c.keyGetPlace(req)
	return c.cachedFetchPlace(ctx, key, func() (*v1.Place, error) {
		return c.inner.GetPlace(ctx, req)
	})
}

func (c *CachedPlaces) SearchText(ctx context.Context, req *v1.SearchTextRequest) (*v1.SearchTextResponse, error) {
	key := c.keySearchText(req)
	return c.cachedFetchSearchResponse(ctx, key, func() (*v1.SearchTextResponse, error) {
		return c.inner.SearchText(ctx, req)
	})
}

func (c *CachedPlaces) SearchRestaurants(ctx context.Context, req *v1.SearchRestaurantsRequest) (*v1.SearchTextResponse, error) {
	key := c.keySearchRestaurants(req)
	return c.cachedFetchSearchResponse(ctx, key, func() (*v1.SearchTextResponse, error) {
		return c.inner.SearchRestaurants(ctx, req)
	})
}