// Package database implements wrapper and utility methods for manipulating
// database elements
package database

import (
	"encoding/json"
	"math/rand"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

type Tag struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

var tagColors = []string{
	"blue",
	"purple",
	"cyan",
	"green",
	"yellow",
	"orange",
	"red",
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

func ParseTags(input string) []Tag {
	cleaned := strings.ReplaceAll(input, ",", " ")

	fields := strings.Fields(cleaned)

	var tags []Tag
	for _, f := range fields {
		tags = append(tags, Tag{
			Name:  f,
			Color: randomColor(),
		})
	}

	return tags
}

func TagsToCommaList(jsonInput string) string {
	tags := JSONToTags(jsonInput)
	if len(tags) == 0 {
		return ""
	}

	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Name
	}

	return strings.Join(names, ", ")
}

func TagsToJSON(tags []Tag) string {
	b, err := json.Marshal(tags)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func randomColor() string {
	return tagColors[rand.New(
		rand.NewSource(
			time.Now().UnixNano(),
		),
	).Intn(
		len(tagColors),
	)]
}
