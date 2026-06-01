package sites

import (
	"errors"
	"reflect"
	"testing"
)

func TestNormalizeTagsRemovesLeadingHash(t *testing.T) {
	tags, err := normalizeTags([]string{"#mytag", "other", " #spaced "})
	if err != nil {
		t.Fatalf("normalizeTags returned error: %v", err)
	}

	want := []string{"mytag", "other", "spaced"}
	if !reflect.DeepEqual(tags, want) {
		t.Fatalf("normalizeTags() = %v, want %v", tags, want)
	}
}

func TestNormalizeTagsRejectsOnlyHash(t *testing.T) {
	_, err := normalizeTags([]string{"#"})
	if !errors.Is(err, ErrInvalidTags) {
		t.Fatalf("normalizeTags() error = %v, want %v", err, ErrInvalidTags)
	}
}
