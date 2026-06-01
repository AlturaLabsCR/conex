// Package database defines the application's database interfaces.
package database

import "context"

type AccountLoginRequest struct {
	Email     string
	Otp       int64
	ExpiresAt int64
}

type Database interface {
	Querier() Querier
	WithTx(ctx context.Context, fn func(q Querier) error) (err error)
	Exec(ctx context.Context, sql string) (err error)
	IsErrNotFound(err error) bool
	Close(ctx context.Context) (err error)
}

type Querier interface {
	// OncesertAccountByEmail creates an account by email if not yet exists.
	// It returns the email's subject.
	OncesertAccountByEmail(ctx context.Context, email string, createdAt int64) (sub int64, err error)

	// UpdateAccountEmail updates the email for the account subject.
	UpdateAccountEmail(ctx context.Context, sub int64, email string) error

	// DeleteAccount deletes the account for the account subject.
	DeleteAccount(ctx context.Context, sub int64) error

	// SelectAccountBySub returns the account for the account subject.
	SelectAccountBySub(ctx context.Context, sub int64) (*Account, error)

	// SelectAccountSubscriptionBySub returns the account's subscription and plan.
	SelectAccountSubscriptionBySub(ctx context.Context, sub int64) (*AccountSubscription, error)

	// SelectPlans returns available plans.
	SelectPlans(ctx context.Context) ([]Plan, error)

	// CreatePayment records a pending payment order.
	CreatePayment(ctx context.Context, orderID string, sub int64, planID string, amount int64, currency string, status string) error

	// SelectPaymentByOrderID returns a payment by PayPal order ID.
	SelectPaymentByOrderID(ctx context.Context, orderID string) (*Payment, error)

	// CapturePayment marks a payment order captured for an account subject.
	CapturePayment(ctx context.Context, orderID string, sub int64) (*Payment, error)

	// UpdateSubscriptionPlan updates the account subscription after a captured payment.
	UpdateSubscriptionPlan(ctx context.Context, sub int64, planID string, dueDate string) error

	// CreateSite creates a site for an account subject.
	CreateSite(ctx context.Context, sub int64, path string, public bool) error

	// CreateSiteTimestamps creates the creation and last-modified timestamps for a site.
	CreateSiteTimestamps(ctx context.Context, path string) error

	// CreateSiteClicks creates the click counter for a site.
	CreateSiteClicks(ctx context.Context, path string) error

	// UpsertSiteName creates or updates the display name for a site.
	UpsertSiteName(ctx context.Context, path string, name string) error

	// DeleteSiteTags deletes all tags for a site.
	DeleteSiteTags(ctx context.Context, path string) error

	// CreateSiteTag creates a tag for a site.
	CreateSiteTag(ctx context.Context, path string, tag string) error

	// SelectSiteByPath returns the site for the site path.
	SelectSiteByPath(ctx context.Context, path string) (*Site, error)

	// SelectSitesBySub returns the sites owned by the account subject.
	SelectSitesBySub(ctx context.Context, sub int64) ([]Site, error)

	// SelectPublicSitesByClicks returns public sites ordered by click count.
	SelectPublicSitesByClicks(ctx context.Context, limit int64, offset int64) ([]Site, error)

	// SelectPublicSitesByCreatedAt returns public sites ordered by creation time.
	SelectPublicSitesByCreatedAt(ctx context.Context, limit int64, offset int64) ([]Site, error)

	// SearchPublicSites returns public sites matching a name or tag query.
	SearchPublicSites(ctx context.Context, query string, limit int64, offset int64) ([]Site, error)

	// UpdateSitePublic updates the public status for a site owned by the account subject.
	UpdateSitePublic(ctx context.Context, sub int64, path string, public bool) error

	// IncrementSiteClicks increments a site's click counter by one.
	IncrementSiteClicks(ctx context.Context, path string) error

	// UpdateSiteLastModified updates a site's last modified timestamp to the database current time.
	UpdateSiteLastModified(ctx context.Context, path string) error

	// DeleteSite deletes a site owned by the account subject.
	DeleteSite(ctx context.Context, sub int64, path string) (*Site, error)

	// UpsertAccountEmailChangeRequest creates or updates the pending email change request for an account.
	UpsertAccountEmailChangeRequest(ctx context.Context, sub int64, email string, otp int64, expiresAt int64) error

	// SelectAccountEmailChangeRequestBySub returns the pending email change request for an account.
	SelectAccountEmailChangeRequestBySub(ctx context.Context, sub int64) (*AccountEmailChangeRequest, error)

	// DeleteAccountEmailChangeRequest deletes the pending email change request for an account.
	DeleteAccountEmailChangeRequest(ctx context.Context, sub int64) error

	// UpsertAccountLoginRequest creates or updates the login request for an email.
	UpsertAccountLoginRequest(ctx context.Context, email string, otp int64, expiresAt int64) error

	// SelectAccountLoginRequest returns the login request for an email.
	SelectAccountLoginRequest(ctx context.Context, email string) (*AccountLoginRequest, error)

	// DeleteAccountLoginRequest deletes the login request for an email.
	DeleteAccountLoginRequest(ctx context.Context, email string) error
}
