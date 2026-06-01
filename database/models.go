package database

type Account struct {
	Sub       int64
	Email     string
	CreatedAt int64
}

type AccountSubscription struct {
	Sub             int64
	PlanID          string
	PlanNameKey     string
	PriceAmount     int64
	PriceCurrency   string
	BillingUnit     string
	BillingCount    int64
	SupportsRenewal bool
	Policy          PlanPolicy
	Status          string
	DueDate         string
}

type PlanPolicy struct {
	MaxBytesPerSite    int64 `json:"max_bytes_per_site"`
	MaxSites           int64 `json:"max_sites"`
	MaxSubpathsPerSite int64 `json:"max_subpaths_per_site"`
}

type Plan struct {
	ID              string
	NameKey         string
	PriceAmount     int64
	PriceCurrency   string
	BillingUnit     string
	BillingCount    int64
	SupportsRenewal bool
	Policy          PlanPolicy
}

type Payment struct {
	OrderID   string
	Sub       int64
	PlanID    string
	Amount    int64
	Currency  string
	Status    string
	CreatedAt int64
	PaidAt    int64
}

type AccountEmailChangeRequest struct {
	Sub       int64
	Email     string
	Otp       int64
	ExpiresAt int64
}

type Site struct {
	Sub          int64    `json:"sub"`
	Path         string   `json:"path"`
	Public       bool     `json:"public"`
	Name         string   `json:"name"`
	Tags         []string `json:"tags"`
	CreatedAt    int64    `json:"created_at"`
	LastModified int64    `json:"last_modified"`
	Clicks       int64    `json:"clicks"`
}
