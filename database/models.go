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
	Status          string
	DueDate         string
}

type AccountEmailChangeRequest struct {
	Sub       int64
	Email     string
	Otp       int64
	ExpiresAt int64
}

type Site struct {
	Sub    int64    `json:"sub"`
	Path   string   `json:"path"`
	Public bool     `json:"public"`
	Name   string   `json:"name"`
	Tags   []string `json:"tags"`
}
