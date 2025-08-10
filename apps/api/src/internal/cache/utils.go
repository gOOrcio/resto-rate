package cache

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// HashKey creates a hash from multiple string parts
func HashKey(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte{0})
		h.Write([]byte(p))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// NormalizeValue converts different value types to a consistent string representation
func NormalizeValue(v any) string {
	switch val := v.(type) {
	case []string:
		sorted := make([]string, len(val))
		copy(sorted, val)
		sort.Strings(sorted)
		return "[" + strings.Join(sorted, ",") + "]"
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
