package test

import (
	"api/src/internal/utils"
	"testing"
)

func TestRequiredTags_Count(t *testing.T) {
	if got := len(utils.RequiredTags); got != 38 {
		t.Fatalf("expected 38 required tags, got %d", got)
	}
}

func TestRequiredTags_Slugs_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for _, tag := range utils.RequiredTags {
		if seen[tag.Slug] {
			t.Fatalf("duplicate slug: %s", tag.Slug)
		}
		seen[tag.Slug] = true
		if tag.Slug == "" {
			t.Fatal("tag has empty slug")
		}
		if tag.Label == "" {
			t.Fatal("tag has empty label")
		}
		if tag.Category == "" {
			t.Fatal("tag has empty category")
		}
	}
}
