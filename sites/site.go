// Package sites wraps site persistence and object storage operations.
package sites

import (
	"bytes"
	"context"
	"errors"
	"io"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"app/database"
	"github.com/microcosm-cc/bluemonday"
	"github.com/tavocg/go-storage"
)

var (
	ErrInvalidPath     = errors.New("invalid site path")
	ErrInvalidName     = errors.New("invalid site name")
	ErrInvalidTags     = errors.New("invalid site tags")
	ErrPathUnavailable = errors.New("site path unavailable")
	ErrSiteNotFound    = errors.New("site not found")
	ErrSiteSizeLimit   = errors.New("site size limit exceeded")
	ErrSiteCountLimit  = errors.New("site count limit exceeded")
	ErrSubpathLimit    = errors.New("site subpath limit exceeded")
	sitePathPattern    = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$`)
	siteHTMLPolicy     = newSiteHTMLPolicy()
)

const (
	sitePathMinLength = 3
	sitePathMaxLength = 255
	siteNameMaxLength = 255
	siteTagMaxLength  = 64
	siteTagsMaxCount  = 32
	siteListPageSize  = 20
	siteClickWindow   = time.Hour
	siteClickMaxKeys  = 100000
)

type Sites interface {
	Create(ctx context.Context, sub int64, path string, name string, tags []string, html io.Reader) (*database.Site, error)
	Get(ctx context.Context, path string, clickClientHash string) (*database.Site, string, error)
	GetOwned(ctx context.Context, sub int64, path string) (*database.Site, string, error)
	List(ctx context.Context, sub int64) ([]database.Site, error)
	ListTop(ctx context.Context, page int64) ([]database.Site, error)
	ListLatest(ctx context.Context, page int64) ([]database.Site, error)
	Search(ctx context.Context, query string, page int64) ([]database.Site, error)
	Update(ctx context.Context, sub int64, path string, update SiteUpdate) (*database.Site, error)
	Delete(ctx context.Context, sub int64, path string) error
	DeleteAll(ctx context.Context, sub int64) error
	IsPathAvailable(ctx context.Context, path string) (bool, error)
}

type SiteUpdate struct {
	Public *bool
	Name   *string
	Tags   *[]string
	HTML   *string
}

type Service struct {
	db               database.Database
	privateStorage   *storage.Storage
	publicStorage    *storage.Storage
	siteClickMu      sync.Mutex
	siteClickHistory map[string]time.Time
	siteClickOrder   []string
}

var _ Sites = (*Service)(nil)

func New(db database.Database, privateStorage, publicStorage *storage.Storage) *Service {
	if db == nil {
		panic("sites database is required")
	}
	if privateStorage == nil {
		panic("sites private storage is required")
	}
	if publicStorage == nil {
		panic("sites public storage is required")
	}

	return &Service{
		db:               db,
		privateStorage:   privateStorage,
		publicStorage:    publicStorage,
		siteClickHistory: make(map[string]time.Time),
	}
}

func (s *Service) Create(ctx context.Context, sub int64, path string, name string, tags []string, html io.Reader) (*database.Site, error) {
	path, err := normalizePath(path)
	if err != nil {
		return nil, err
	}
	name, err = normalizeName(name)
	if err != nil {
		return nil, err
	}
	tags, err = normalizeTags(tags)
	if err != nil {
		return nil, err
	}

	body, err := readSiteHTML(html)
	if err != nil {
		return nil, err
	}

	policy, err := s.sitePolicy(ctx, sub)
	if err != nil {
		return nil, err
	}
	if err := enforceSiteSizeLimit(policy, body); err != nil {
		return nil, err
	}
	if err := enforceSiteSubpathLimit(policy, path); err != nil {
		return nil, err
	}
	ownedSites, err := s.db.Querier().SelectSitesBySub(ctx, sub)
	if err != nil {
		return nil, err
	}
	if policy.MaxSites > 0 && int64(len(ownedSites)) >= policy.MaxSites {
		return nil, ErrSiteCountLimit
	}

	available, err := s.IsPathAvailable(ctx, path)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, ErrPathUnavailable
	}

	site := &database.Site{
		Sub:    sub,
		Path:   path,
		Public: false,
		Name:   name,
		Tags:   tags,
	}

	err = s.db.WithTx(ctx, func(q database.Querier) error {
		if err := q.CreateSite(ctx, sub, path, false); err != nil {
			return err
		}
		if err := q.CreateSiteTimestamps(ctx, path); err != nil {
			return err
		}
		if err := q.CreateSiteClicks(ctx, path); err != nil {
			return err
		}
		if err := q.UpsertSiteName(ctx, path, name); err != nil {
			return err
		}
		for _, tag := range tags {
			if err := q.CreateSiteTag(ctx, path, tag); err != nil {
				return err
			}
		}

		_, err := s.privateStorage.Put(
			ctx,
			bytes.NewReader(body),
			storage.WithKey(path),
			storage.WithSizeLimit(int64(len(body))),
			storage.WithContentType("text/html"),
		)
		return err
	})
	if err != nil {
		if _, selectErr := s.db.Querier().SelectSiteByPath(ctx, path); selectErr == nil {
			return nil, ErrPathUnavailable
		}

		return nil, err
	}

	site, err = s.db.Querier().SelectSiteByPath(ctx, path)
	if err != nil {
		return nil, err
	}

	return site, nil
}

func (s *Service) Get(ctx context.Context, path string, clickClientHash string) (*database.Site, string, error) {
	path, err := normalizePath(path)
	if err != nil {
		return nil, "", err
	}

	site, err := s.db.Querier().SelectSiteByPath(ctx, path)
	if err != nil {
		if s.db.IsErrNotFound(err) {
			return nil, "", ErrSiteNotFound
		}

		return nil, "", err
	}
	if !site.Public {
		return nil, "", ErrSiteNotFound
	}

	body, err := s.publicStorage.Get(ctx, &storage.ObjectHead{Key: path})
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotFound) {
			return nil, "", ErrSiteNotFound
		}

		return nil, "", err
	}
	defer func() {
		_ = body.Close()
	}()

	html, err := io.ReadAll(body)
	if err != nil {
		return nil, "", err
	}

	if err := s.countClick(ctx, path, clickClientHash); err != nil {
		return nil, "", err
	}

	return site, string(html), nil
}

func (s *Service) countClick(ctx context.Context, path string, clickClientHash string) error {
	if clickClientHash != "" && !s.claimClick(path, clickClientHash, time.Now()) {
		return nil
	}

	return s.db.Querier().IncrementSiteClicks(ctx, path)
}

func (s *Service) claimClick(path string, clickClientHash string, now time.Time) bool {
	key := path + "\x00" + clickClientHash

	s.siteClickMu.Lock()
	defer s.siteClickMu.Unlock()

	s.pruneSiteClickHistoryLocked(now)
	if _, ok := s.siteClickHistory[key]; ok {
		return false
	}
	if len(s.siteClickHistory) >= siteClickMaxKeys {
		s.evictOldestSiteClickLocked()
	}

	s.siteClickHistory[key] = now
	s.siteClickOrder = append(s.siteClickOrder, key)

	return true
}

func (s *Service) pruneSiteClickHistoryLocked(now time.Time) {
	kept := s.siteClickOrder[:0]
	for _, key := range s.siteClickOrder {
		lastClicked, ok := s.siteClickHistory[key]
		if !ok {
			continue
		}
		if now.Sub(lastClicked) >= siteClickWindow {
			delete(s.siteClickHistory, key)
			continue
		}

		kept = append(kept, key)
	}

	s.siteClickOrder = kept
}

func (s *Service) evictOldestSiteClickLocked() {
	for len(s.siteClickOrder) > 0 {
		key := s.siteClickOrder[0]
		s.siteClickOrder = s.siteClickOrder[1:]
		if _, ok := s.siteClickHistory[key]; ok {
			delete(s.siteClickHistory, key)
			return
		}
	}
}

func (s *Service) List(ctx context.Context, sub int64) ([]database.Site, error) {
	return s.db.Querier().SelectSitesBySub(ctx, sub)
}

func (s *Service) ListTop(ctx context.Context, page int64) ([]database.Site, error) {
	return s.db.Querier().SelectPublicSitesByClicks(ctx, siteListPageSize, siteListOffset(page))
}

func (s *Service) ListLatest(ctx context.Context, page int64) ([]database.Site, error) {
	return s.db.Querier().SelectPublicSitesByCreatedAt(ctx, siteListPageSize, siteListOffset(page))
}

func (s *Service) Search(ctx context.Context, query string, page int64) ([]database.Site, error) {
	query = strings.TrimSpace(query)
	if query == "" || len(query) > siteNameMaxLength {
		return nil, ErrInvalidName
	}

	return s.db.Querier().SearchPublicSites(ctx, query, siteListPageSize, siteListOffset(page))
}

func siteListOffset(page int64) int64 {
	if page < 1 {
		page = 1
	}

	return (page - 1) * siteListPageSize
}

func (s *Service) GetOwned(ctx context.Context, sub int64, path string) (*database.Site, string, error) {
	path, err := normalizePath(path)
	if err != nil {
		return nil, "", err
	}

	site, err := s.db.Querier().SelectSiteByPath(ctx, path)
	if err != nil {
		if s.db.IsErrNotFound(err) {
			return nil, "", ErrSiteNotFound
		}

		return nil, "", err
	}
	if site.Sub != sub {
		return nil, "", ErrSiteNotFound
	}

	source := s.privateStorage
	if site.Public {
		source = s.publicStorage
	}

	body, err := source.Get(ctx, &storage.ObjectHead{Key: path})
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotFound) {
			return nil, "", ErrSiteNotFound
		}

		return nil, "", err
	}
	defer func() {
		_ = body.Close()
	}()

	html, err := io.ReadAll(body)
	if err != nil {
		return nil, "", err
	}

	return site, string(html), nil
}

func (s *Service) Update(ctx context.Context, sub int64, path string, update SiteUpdate) (*database.Site, error) {
	path, err := normalizePath(path)
	if err != nil {
		return nil, err
	}
	var name string
	if update.Name != nil {
		name, err = normalizeName(*update.Name)
		if err != nil {
			return nil, err
		}
	}
	var tags []string
	if update.Tags != nil {
		tags, err = normalizeTags(*update.Tags)
		if err != nil {
			return nil, err
		}
	}
	var html []byte
	if update.HTML != nil {
		html, err = sanitizeSiteHTML([]byte(*update.HTML))
		if err != nil {
			return nil, err
		}
	}

	current, err := s.db.Querier().SelectSiteByPath(ctx, path)
	if err != nil {
		if s.db.IsErrNotFound(err) {
			return nil, ErrSiteNotFound
		}

		return nil, err
	}
	if current.Sub != sub {
		return nil, ErrSiteNotFound
	}

	if update.HTML != nil {
		policy, err := s.sitePolicy(ctx, sub)
		if err != nil {
			return nil, err
		}
		if err := enforceSiteSizeLimit(policy, html); err != nil {
			return nil, err
		}
	}

	// TODO: Shouldn't this else if be a separate if, in case the request has both html AND public status?
	if update.HTML != nil {
		destination := s.privateStorage
		if current.Public {
			destination = s.publicStorage
		}
		if update.Public != nil && *update.Public {
			destination = s.publicStorage
		}
		if update.Public != nil && !*update.Public {
			destination = s.privateStorage
		}

		if _, err := destination.Put(
			ctx,
			bytes.NewReader(html),
			storage.WithKey(path),
			storage.WithSizeLimit(int64(len(html))),
			storage.WithContentType("text/html"),
		); err != nil {
			return nil, err
		}
	} else if update.Public != nil && current.Public != *update.Public {
		source, destination := s.privateStorage, s.publicStorage
		if !*update.Public {
			source, destination = s.publicStorage, s.privateStorage
		}

		size, ok := source.List()[path]
		if !ok {
			return nil, ErrSiteNotFound
		}

		body, err := source.Get(ctx, &storage.ObjectHead{Key: path})
		if err != nil {
			if errors.Is(err, storage.ErrObjectNotFound) {
				return nil, ErrSiteNotFound
			}

			return nil, err
		}
		defer func() {
			_ = body.Close()
		}()

		if _, err := destination.Put(
			ctx,
			body,
			storage.WithKey(path),
			storage.WithSizeLimit(size),
			storage.WithContentType("text/html"),
		); err != nil {
			return nil, err
		}
	}

	if err := s.db.WithTx(ctx, func(q database.Querier) error {
		if update.Public != nil && current.Public != *update.Public {
			if err := q.UpdateSitePublic(ctx, sub, path, *update.Public); err != nil {
				return err
			}
		}
		if update.Name != nil {
			if err := q.UpsertSiteName(ctx, path, name); err != nil {
				return err
			}
		}
		if update.Tags != nil {
			if err := q.DeleteSiteTags(ctx, path); err != nil {
				return err
			}
			for _, tag := range tags {
				if err := q.CreateSiteTag(ctx, path, tag); err != nil {
					return err
				}
			}
		}
		if siteUpdateTouchesSite(current, update) {
			if err := q.UpdateSiteLastModified(ctx, path); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if update.Public != nil && current.Public != *update.Public {
		source := s.privateStorage
		if !*update.Public {
			source = s.publicStorage
		}
		if err := source.Delete(ctx, &storage.ObjectHead{Key: path}); err != nil && !errors.Is(err, storage.ErrObjectNotFound) {
			return nil, err
		}
	}

	// TODO: Is this necessary?
	site, err := s.db.Querier().SelectSiteByPath(ctx, path)
	if err != nil {
		return nil, err
	}

	return site, nil
}

func (s *Service) Delete(ctx context.Context, sub int64, path string) error {
	path, err := normalizePath(path)
	if err != nil {
		return err
	}

	return s.db.WithTx(ctx, func(q database.Querier) error {
		site, err := q.SelectSiteByPath(ctx, path)
		if err != nil {
			if s.db.IsErrNotFound(err) {
				return ErrSiteNotFound
			}

			return err
		}
		if site.Sub != sub {
			return ErrSiteNotFound
		}

		if err := s.privateStorage.Delete(ctx, &storage.ObjectHead{Key: path}); err != nil && !errors.Is(err, storage.ErrObjectNotFound) {
			return err
		}
		if err := s.publicStorage.Delete(ctx, &storage.ObjectHead{Key: path}); err != nil && !errors.Is(err, storage.ErrObjectNotFound) {
			return err
		}

		if _, err := q.DeleteSite(ctx, sub, path); err != nil {
			if s.db.IsErrNotFound(err) {
				return ErrSiteNotFound
			}

			return err
		}

		return nil
	})
}

func (s *Service) DeleteAll(ctx context.Context, sub int64) error {
	ownedSites, err := s.db.Querier().SelectSitesBySub(ctx, sub)
	if err != nil {
		return err
	}

	for _, site := range ownedSites {
		if err := s.Delete(ctx, sub, site.Path); err != nil && !errors.Is(err, ErrSiteNotFound) {
			return err
		}
	}

	return nil
}

func (s *Service) IsPathAvailable(ctx context.Context, path string) (bool, error) {
	path, err := normalizePath(path)
	if err != nil {
		return false, err
	}

	if s.privateStorage.Exists(path) || s.publicStorage.Exists(path) {
		return false, nil
	}

	_, err = s.db.Querier().SelectSiteByPath(ctx, path)
	if err == nil {
		return false, nil
	}
	if s.db.IsErrNotFound(err) {
		return true, nil
	}

	return false, err
}

func siteUpdateTouchesSite(current *database.Site, update SiteUpdate) bool {
	return update.HTML != nil ||
		update.Name != nil ||
		update.Tags != nil ||
		(update.Public != nil && current.Public != *update.Public)
}

func (s *Service) sitePolicy(ctx context.Context, sub int64) (database.PlanPolicy, error) {
	subscription, err := s.db.Querier().SelectAccountSubscriptionBySub(ctx, sub)
	if err != nil {
		return database.PlanPolicy{}, err
	}

	return subscription.Policy, nil
}

func enforceSiteSizeLimit(policy database.PlanPolicy, body []byte) error {
	if policy.MaxBytesPerSite > 0 && int64(len(body)) > policy.MaxBytesPerSite {
		return ErrSiteSizeLimit
	}

	return nil
}

func enforceSiteSubpathLimit(policy database.PlanPolicy, path string) error {
	if policy.MaxSubpathsPerSite >= 0 && countSiteSubpaths(path) > policy.MaxSubpathsPerSite {
		return ErrSubpathLimit
	}

	return nil
}

func countSiteSubpaths(path string) int64 {
	path = strings.Trim(path, "/")
	if path == "" || !strings.Contains(path, "/") {
		return 0
	}

	return int64(strings.Count(path, "/"))
}

func readSiteHTML(reader io.Reader) ([]byte, error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return sanitizeSiteHTML(body)
}

func sanitizeSiteHTML(body []byte) ([]byte, error) {
	body = siteHTMLPolicy.SanitizeBytes(body)

	return body, nil
}

func normalizePath(path string) (string, error) {
	path = strings.TrimSpace(strings.ToLower(path))
	if len(path) < sitePathMinLength || len(path) > sitePathMaxLength {
		return "", ErrInvalidPath
	}
	if !sitePathPattern.MatchString(path) {
		return "", ErrInvalidPath
	}

	return path, nil
}

func normalizeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if len(name) == 0 || len(name) > siteNameMaxLength {
		return "", ErrInvalidName
	}

	return name, nil
}

func normalizeTags(tags []string) ([]string, error) {
	if len(tags) > siteTagsMaxCount {
		return nil, ErrInvalidTags
	}

	out := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		tag = strings.TrimPrefix(tag, "#")
		if len(tag) == 0 || len(tag) > siteTagMaxLength {
			return nil, ErrInvalidTags
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	slices.Sort(out)

	return out, nil
}

func newSiteHTMLPolicy() *bluemonday.Policy {
	policy := bluemonday.UGCPolicy()
	policy.RequireNoReferrerOnLinks(true)
	policy.AllowAttrs("class").Matching(bluemonday.SpaceSeparatedTokens).Globally()
	policy.AllowDataURIImages()

	return policy
}
