// Package sites wraps site persistence and object storage operations.
package sites

import (
	"bytes"
	"context"
	"errors"
	"io"
	"regexp"
	"strings"

	"app/database"
	"github.com/microcosm-cc/bluemonday"
	"github.com/tavocg/go-storage"
)

var (
	ErrInvalidPath     = errors.New("invalid site path")
	ErrPathUnavailable = errors.New("site path unavailable")
	ErrSiteNotFound    = errors.New("site not found")
	sitePathPattern    = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$`)
	siteHTMLPolicy     = newSiteHTMLPolicy()
)

const (
	sitePathMinLength = 3
	sitePathMaxLength = 255
)

type Sites interface {
	Create(ctx context.Context, sub int64, path string, html io.Reader) (*database.Site, error)
	Get(ctx context.Context, path string) (*database.Site, string, error)
	GetOwned(ctx context.Context, sub int64, path string) (*database.Site, string, error)
	List(ctx context.Context, sub int64) ([]database.Site, error)
	SetPublic(ctx context.Context, sub int64, path string, public bool) (*database.Site, error)
	Delete(ctx context.Context, sub int64, path string) error
	DeleteAll(ctx context.Context, sub int64) error
	IsPathAvailable(ctx context.Context, path string) (bool, error)
}

type Service struct {
	db             database.Database
	privateStorage *storage.Storage
	publicStorage  *storage.Storage
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
		db:             db,
		privateStorage: privateStorage,
		publicStorage:  publicStorage,
	}
}

func (s *Service) Create(ctx context.Context, sub int64, path string, html io.Reader) (*database.Site, error) {
	path, err := normalizePath(path)
	if err != nil {
		return nil, err
	}

	available, err := s.IsPathAvailable(ctx, path)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, ErrPathUnavailable
	}

	body, err := io.ReadAll(html)
	if err != nil {
		return nil, err
	}
	body = siteHTMLPolicy.SanitizeBytes(body)

	site := &database.Site{
		Sub:    sub,
		Path:   path,
		Public: false,
	}

	err = s.db.WithTx(ctx, func(q database.Querier) error {
		if err := q.CreateSite(ctx, sub, path, false); err != nil {
			return err
		}

		_, err := s.privateStorage.Put(
			ctx,
			bytes.NewReader(body),
			storage.WithKey(path),
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

	return site, nil
}

func (s *Service) Get(ctx context.Context, path string) (*database.Site, string, error) {
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

	return site, string(html), nil
}

func (s *Service) List(ctx context.Context, sub int64) ([]database.Site, error) {
	return s.db.Querier().SelectSitesBySub(ctx, sub)
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

func (s *Service) SetPublic(ctx context.Context, sub int64, path string, public bool) (*database.Site, error) {
	path, err := normalizePath(path)
	if err != nil {
		return nil, err
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
	if current.Public == public {
		return current, nil
	}

	source, destination := s.privateStorage, s.publicStorage
	if !public {
		source, destination = s.publicStorage, s.privateStorage
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
		storage.WithContentType("text/html"),
	); err != nil {
		return nil, err
	}

	site, err := s.db.Querier().UpdateSitePublic(ctx, sub, path, public)
	if err != nil {
		return nil, err
	}

	if err := source.Delete(ctx, &storage.ObjectHead{Key: path}); err != nil && !errors.Is(err, storage.ErrObjectNotFound) {
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

func newSiteHTMLPolicy() *bluemonday.Policy {
	policy := bluemonday.UGCPolicy()
	policy.RequireNoReferrerOnLinks(true)
	policy.AllowAttrs("class").Matching(bluemonday.SpaceSeparatedTokens).Globally()
	policy.AllowDataURIImages()

	return policy
}
