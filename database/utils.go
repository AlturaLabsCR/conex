// Package database implements wrapper and utility methods for manipulating
// database elements
package database

import (
	"encoding/json"

	"github.com/microcosm-cc/bluemonday"
)

type Tag struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func JSONToTags(s string) []Tag {
	var tags []Tag

	err := json.Unmarshal([]byte(s), &tags)
	if err != nil {
		return []Tag{}
	}

	return tags
}

func SanitizeHTML(s string) string {
	p := bluemonday.UGCPolicy()
	return p.Sanitize(s)
}
