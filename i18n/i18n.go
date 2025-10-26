// Package i18n
package i18n

import (
	"net/http"

	"golang.org/x/text/language"
)

// Locales is a map of locales and their translations.
// Example: { "en": {"hello": "Hello"}, "es": {"hello": "Hola"} }
type Locales map[string]map[string]string

// Translator holds all available translations.
type Translator struct {
	locales Locales
}

// New creates a new Translator with the given locales.
func New(locales Locales) *Translator {
	return &Translator{locales: locales}
}

// Translate translates a given key for a specific locale.
// Fallback: returns the key itself if not found.
func (t *Translator) Translate(locale, key string) string {
	if loc, ok := t.locales[locale]; ok {
		if val, ok := loc[key]; ok {
			return val
		}
	}
	return key
}

// TranslateHTTPRequest auto-detects the user's preferred language
// (from cookie or Accept-Language header) and returns a closure
// that translates by key, e.g. tr("hello").
func (t *Translator) TranslateHTTPRequest(r *http.Request) func(string) string {
	// Infer supported locales from the Translator
	supported := make([]string, 0, len(t.locales))
	for lang := range t.locales {
		supported = append(supported, lang)
	}

	// Fallback = first locale
	fallback := ""
	if len(supported) > 0 {
		fallback = supported[0]
	}

	// Detect user language
	lang := DetectLanguage(r, supported, fallback)

	return func(key string) string {
		return t.Translate(lang, key)
	}
}

// DetectLanguage picks a language from cookie or header.
func DetectLanguage(r *http.Request, supported []string, fallback string) string {
	if c, err := r.Cookie("lang"); err == nil {
		for _, s := range supported {
			if s == c.Value {
				return s
			}
		}
	}
	return DetectLanguageFromHeader(r, supported, fallback)
}

// DetectLanguageFromHeader reads and parses Accept-Language.
func DetectLanguageFromHeader(r *http.Request, supported []string, fallback string) string {
	header := r.Header.Get("Accept-Language")
	if header == "" {
		return fallback
	}

	tags, _, err := language.ParseAcceptLanguage(header)
	if err != nil {
		return fallback
	}

	matcher := language.NewMatcher(localesToTags(supported))
	tag, _, _ := matcher.Match(tags...)
	base, _ := tag.Base()
	return base.String()
}

func localesToTags(locales []string) []language.Tag {
	tags := make([]language.Tag, len(locales))
	for i, l := range locales {
		tags[i] = language.Make(l)
	}
	return tags
}
