package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"app/config"
	"app/internal/db"
)

const (
	tokenEndpoint               = "v1/oauth2/token"
	createOrderEndpoint         = "v2/checkout/orders"
	captureOrderEndpoint        = "v2/checkout/orders"
	createSubscriptionEndpoint  = "v1/billing/subscriptions"
	captureSubscriptionEndpoint = "v1/billing/subscriptions"

	intentCapture = "capture"
)

var ErrRespondCapture error = errors.New("error responding")

type AccessToken struct {
	Scope     string `json:"scope"`
	Token     string `json:"access_token"`
	Type      string `json:"token_type"`
	AppID     string `json:"app_id"`
	ExpiresIn int64  `json:"expires_in"`
	Nonce     string `json:"nonce"`
}

type SubscriptionResponse struct {
	ID               string          `json:"id,omitempty"`
	Status           string          `json:"status,omitempty"`
	StatusUpdateTime time.Time       `json:"status_update_time"`
	PlanID           string          `json:"plan_id,omitempty"`
	PlanOverridden   bool            `json:"plan_overridden,omitempty"`
	StartTime        time.Time       `json:"start_time"`
	Quantity         string          `json:"quantity,omitempty"`
	ShippingAmount   *ShippingAmount `json:"shipping_amount"`
	Subscriber       *Subscriber     `json:"subscriber"`
	CreateTime       time.Time       `json:"create_time"`
	Links            []Links         `json:"links,omitempty"`
}

type SubscriptionRequest struct {
	PlanID             string              `json:"plan_id,omitempty"`
	StartTime          time.Time           `json:"start_time"`
	Quantity           string              `json:"quantity,omitempty"`
	ShippingAmount     *ShippingAmount     `json:"shipping_amount"`
	Subscriber         *Subscriber         `json:"subscriber"`
	ApplicationContext *ApplicationContext `json:"application_context"`
}

type ApplicationContext struct {
	BrandName          string         `json:"brand_name,omitempty"`
	Locale             string         `json:"locale,omitempty"`
	ShippingPreference string         `json:"shipping_preference,omitempty"`
	UserAction         string         `json:"user_action,omitempty"`
	PaymentMethod      *PaymentMethod `json:"payment_method"`
	ReturnURL          string         `json:"return_url,omitempty"`
	CancelURL          string         `json:"cancel_url,omitempty"`
}

type PaymentMethod struct {
	PayerSelected  string `json:"payer_selected,omitempty"`
	PayeePreferred string `json:"payee_preferred,omitempty"`
}

type ShippingAmount struct {
	CurrencyCode string `json:"currency_code,omitempty"`
	Value        string `json:"value,omitempty"`
}

type Subscriber struct {
	Name            *Name            `json:"name"`
	EmailAddress    string           `json:"email_address,omitempty"`
	ShippingAddress *ShippingAddress `json:"shipping_address"`

	// subscription response
	PayerID string `json:"payer_id,omitempty"`
}

type ShippingAddress struct {
	Name    *Name    `json:"name"`
	Address *Address `json:"address,omitempty"`
}

type CompleteOrderResponse struct {
	ID            string          `json:"id"`
	Status        string          `json:"status"`
	PaymentSource *PaymentSource  `json:"payment_source"`
	PurchaseUnits []PurchaseUnits `json:"purchase_units"`
	Payer         *Payer          `json:"payer"`
	Links         []Links         `json:"links,omitempty"`
}

type Name struct {
	GivenName string `json:"given_name,omitempty"`
	SurName   string `json:"surname,omitempty"`
	FullName  string `json:"full_name,omitempty"`
}

type Payer struct {
	Name         *Name  `json:"name"`
	EmailAddress string `json:"email_address"`
	PayerID      string `json:"payer_id"`
}

type OrderResponse struct {
	ID            string         `json:"id"`
	Status        string         `json:"status"`
	PaymentSource *PaymentSource `json:"payment_source"`
	Links         []Links        `json:"links,omitempty"`
}

type Links struct {
	Href   string `json:"href,omitempty"`
	Rel    string `json:"rel,omitempty"`
	Method string `json:"method,omitempty"`
}

type OrderRequest struct {
	Intent        string          `json:"intent"`
	PaymentSource *PaymentSource  `json:"payment_source"`
	PurchaseUnits []PurchaseUnits `json:"purchase_units"`
}

type ExperienceContext struct {
	PaymentMethodPreference string `json:"payment_method_preference,omitempty"`
	LandingPage             string `json:"landing_page,omitempty"`
	ShippingPreference      string `json:"shipping_preference,omitempty"`
	UserAction              string `json:"user_action,omitempty"`
	ReturnURL               string `json:"return_url,omitempty"`
	CancelURL               string `json:"cancel_url,omitempty"`
}

type PayPal struct {
	// create order
	ExperienceContext *ExperienceContext `json:"experience_context"`
	Address           *Address           `json:"address"`

	// complete order
	Name          *Name  `json:"name,omitempty"`
	EmailAddress  string `json:"email_address,omitempty"`
	AccountID     string `json:"account_id,omitempty"`
	AccountStatus string `json:"account_status,omitempty"`
}

type PaymentSource struct {
	PayPal *PayPal `json:"paypal"`
}

type ItemTotal struct {
	CurrencyCode string `json:"currency_code,omitempty"`
	Value        string `json:"value,omitempty"`
}

type Address struct {
	AddressLine1 string `json:"address_line_1,omitempty"`
	AddressLine2 string `json:"address_line_2,omitempty"`
	AdminArea2   string `json:"admin_area_2,omitempty"`
	AdminArea1   string `json:"admin_area_1,omitempty"`
	PostalCode   string `json:"postal_code,omitempty"`
	CountryCode  string `json:"country_code,omitempty"`
}

type Shipping struct {
	CurrencyCode string   `json:"currency_code,omitempty"`
	Value        string   `json:"value,omitempty"`
	Address      *Address `json:"address"`
}

type Breakdown struct {
	ItemTotal *ItemTotal `json:"item_total"`
	Shipping  *Shipping  `json:"shipping"`
}

type Amount struct {
	CurrencyCode string     `json:"currency_code"`
	Value        string     `json:"value"`
	Breakdown    *Breakdown `json:"breakdown"`
}

type UnitAmount struct {
	CurrencyCode string `json:"currency_code,omitempty"`
	Value        string `json:"value,omitempty"`
}

type UPC struct {
	Type string `json:"type,omitempty"`
	Code string `json:"code,omitempty"`
}

type Item struct {
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	UnitAmount  *UnitAmount `json:"unit_amount"`
	Quantity    string      `json:"quantity,omitempty"`
	Category    string      `json:"category,omitempty"`
	SKU         string      `json:"sku,omitempty"`
	ImageURL    string      `json:"image_url,omitempty"`
	URL         string      `json:"url,omitempty"`
	UPC         *UPC        `json:"upc"`
}

type SellerProtection struct {
	Status            string   `json:"status,omitempty"`
	DisputeCategories []string `json:"dispute_categories,omitempty"`
}

type GrossAmount struct {
	CurrencyCode string `json:"currency_code,omitempty"`
	Value        string `json:"value,omitempty"`
}

type PaypalFee struct {
	CurrencyCode string `json:"currency_code,omitempty"`
	Value        string `json:"value,omitempty"`
}

type NetAmount struct {
	CurrencyCode string `json:"currency_code,omitempty"`
	Value        string `json:"value,omitempty"`
}

type SellerReceivableBreakdown struct {
	GrossAmount *GrossAmount `json:"gross_amount"`
	PaypalFee   *PaypalFee   `json:"paypal_fee"`
	NetAmount   *NetAmount   `json:"net_amount"`
}

type Captures struct {
	ID                        string                     `json:"id,omitempty"`
	Status                    string                     `json:"status,omitempty"`
	Amount                    Amount                     `json:"amount"`
	SellerProtection          *SellerProtection          `json:"seller_protection"`
	FinalCapture              bool                       `json:"final_capture,omitempty"`
	DisbursementMode          string                     `json:"disbursement_mode,omitempty"`
	SellerReceivableBreakdown *SellerReceivableBreakdown `json:"seller_receivable_breakdown"`
	CreateTime                time.Time                  `json:"create_time"`
	UpdateTime                time.Time                  `json:"update_time"`
	Links                     []Links                    `json:"links,omitempty"`
}

type Payments struct {
	Captures []Captures `json:"captures,omitempty"`
}

type PurchaseUnits struct {
	// create order
	InvoiceID string  `json:"invoice_id,omitempty"`
	Amount    *Amount `json:"amount"`
	Items     []Item  `json:"items,omitempty"`

	// complete order
	ReferenceID string    `json:"reference_id,omitempty"`
	Shipping    *Shipping `json:"shipping"`
	Payments    *Payments `json:"payments"`
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	ord, err := CreateOrder("USD", config.PayPalPurchaseValueStr)
	if err != nil {
		h.Log().Error("failed to create order", "error", err)
		http.Error(w, "failed to create order", http.StatusInternalServerError)
		return
	}

	minOrd := OrderResponse{
		ID: ord.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(minOrd); err != nil {
		h.Log().Error("failed to encode order", "error", err)
		http.Error(w, "Failed to encode order", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) CompleteOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req struct {
		OrderID string `json:"order_id"`
	}

	session, ok := ctx.Value(ctxSessionKey).(db.Session)
	if !ok {
		h.Log().Error("error retrieving session from ctx")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	queries := db.New(h.DB())

	plan, err := queries.GetPlan(ctx, session.SessionUser)
	if err != nil {
		h.Log().Error("failed to query plan", "error", err)
		http.Error(w, "failed to query plan", http.StatusInternalServerError)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Log().Error("failed to decode order_id", "error", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	resp, err := CompleteOrder(req.OrderID)
	if err != nil {
		if !errors.Is(err, ErrRespondCapture) {
			h.Log().Error("failed to complete order", "error", err)
			http.Error(w, "failed to complete order", http.StatusInternalServerError)
			return
		}
	}

	h.Log().Debug("order captured", "order", resp)

	now := time.Now()

	var success int64 = 0

	if resp.Status == "COMPLETED" {
		success = 1
	}

	if _, err = queries.InsertPayment(ctx, db.InsertPaymentParams{
		PaymentUser:       session.SessionUser,
		PaymentAmount:     config.PayPalPurchaseValue,
		PaymentDateUnix:   now.Unix(),
		PaymentSuccessful: success,
		PaymentReference:  resp.ID + ":" + resp.Status,
	}); err != nil {
		h.Log().Error("failed to complete order", "error", err)
	}

	if err := queries.UpdatePlan(ctx, db.UpdatePlanParams{
		UserPlanDueUnix:      now.Add(24 * time.Hour * 365).Unix(),
		UserPlanModifiedUnix: now.Unix(),
		UserPlanActive:       1,
		UserPlanID:           plan.UserPlanID,
	}); err != nil {
		h.Log().Error("failed to update plan", "error", err)
		http.Error(w, "failed to update plan", http.StatusInternalServerError)
		return
	}

	minResp := CompleteOrderResponse{
		ID: resp.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(minResp); err != nil {
		h.Log().Error("failed to encode order", "error", err)
		http.Error(w, "Failed to encode order", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) CompleteSubscription(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SubscriptionID string `json:"subscription_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Log().Error("failed to decode order_id", "error", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	status, err := CompleteSubscription(req.SubscriptionID)
	if err != nil {
		h.Log().Error("failed to capture subscription", "error", err)
		http.Error(w, "failed to capture subscription", http.StatusInternalServerError)
		return
	}

	var empty struct{}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(empty); err != nil {
		h.Log().Error("failed to encode order", "error", err)
		http.Error(w, "Failed to encode order", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)

	// TODO: Handle subscription completed
	// - Assign plan & perms
	// - Redirect to dashboard
}

func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	sub, err := CreateSubscription("P-9SY9600391361462VNEMUJZA")
	if err != nil {
		h.Log().Error("failed to create order", "error", err)
		http.Error(w, "failed to create order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sub); err != nil {
		h.Log().Error("failed to encode order", "error", err)
		http.Error(w, "Failed to encode order", http.StatusInternalServerError)
		return
	}
}

func CreateSubscription(planID string) (SubscriptionResponse, error) {
	resp := SubscriptionResponse{}

	t, err := getAccessToken()
	if err != nil {
		return resp, err
	}
	accessToken := t.Token

	sub := SubscriptionRequest{
		PlanID:    planID,
		StartTime: time.Now().Add(24 * time.Hour),
		ApplicationContext: &ApplicationContext{
			ShippingPreference: "NO_SHIPPING",
		},
	}

	b, err := json.Marshal(sub)
	if err != nil {
		return resp, err
	}

	body := bytes.NewReader(b)
	req, err := http.NewRequest(
		http.MethodPost,
		config.PayPalEndpoint+"/"+createSubscriptionEndpoint,
		body,
	)
	if err != nil {
		return resp, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	raw, err := http.DefaultClient.Do(req)
	if err != nil {
		return resp, err
	}
	defer raw.Body.Close()

	if raw.StatusCode < 200 || raw.StatusCode > 299 {
		body, _ := io.ReadAll(raw.Body)
		return resp, fmt.Errorf("paypal subscription error: [%s] %s", raw.Status, string(body))
	}

	if err := json.NewDecoder(raw.Body).Decode(&resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func CreateOrder(currencyCode, value string) (OrderResponse, error) {
	resp := OrderResponse{}

	t, err := getAccessToken()
	if err != nil {
		return resp, err
	}
	accessToken := t.Token

	order := OrderRequest{
		Intent: strings.ToUpper(intentCapture),
		PurchaseUnits: []PurchaseUnits{
			{
				Amount: &Amount{
					CurrencyCode: currencyCode,
					Value:        value,
				},
			},
		},
		PaymentSource: &PaymentSource{
			PayPal: &PayPal{
				ExperienceContext: &ExperienceContext{
					ShippingPreference: "NO_SHIPPING",
				},
				Address: &Address{
					CountryCode: "CR",
				},
			},
		},
	}

	b, err := json.Marshal(order)
	if err != nil {
		return resp, err
	}

	body := bytes.NewReader(b)
	req, err := http.NewRequest(
		http.MethodPost,
		config.PayPalEndpoint+"/"+createOrderEndpoint,
		body,
	)
	if err != nil {
		return resp, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	raw, err := http.DefaultClient.Do(req)
	if err != nil {
		return resp, err
	}
	defer raw.Body.Close()

	if raw.StatusCode < 200 || raw.StatusCode > 299 {
		body, _ := io.ReadAll(raw.Body)
		return resp, fmt.Errorf("paypal order error: [%s] %s", raw.Status, string(body))
	}

	if err := json.NewDecoder(raw.Body).Decode(&resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func CompleteSubscription(subscriptionID string) (int, error) {
	success := 202

	t, err := getAccessToken()
	if err != nil {
		return http.StatusInternalServerError, err
	}
	accessToken := t.Token

	req, err := http.NewRequest(
		http.MethodPost,
		config.PayPalEndpoint+"/"+captureSubscriptionEndpoint+"/"+subscriptionID+"/"+intentCapture,
		nil,
	)
	if err != nil {
		return http.StatusBadRequest, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	raw, err := http.DefaultClient.Do(req)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer raw.Body.Close()

	if raw.StatusCode != success {
		body, _ := io.ReadAll(raw.Body)
		return http.StatusInternalServerError, fmt.Errorf("paypal order error: [%s] %s", raw.Status, string(body))
	}

	return success, nil
}

func CompleteOrder(orderID string) (CompleteOrderResponse, error) {
	resp := CompleteOrderResponse{}

	t, err := getAccessToken()
	if err != nil {
		return resp, err
	}
	accessToken := t.Token

	req, err := http.NewRequest(
		http.MethodPost,
		config.PayPalEndpoint+"/"+captureOrderEndpoint+"/"+orderID+"/"+intentCapture,
		nil,
	)
	if err != nil {
		return resp, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	raw, err := http.DefaultClient.Do(req)
	if err != nil {
		return resp, err
	}
	defer raw.Body.Close()

	if raw.StatusCode != 200 && raw.StatusCode != 201 {
		body, _ := io.ReadAll(raw.Body)
		return resp, fmt.Errorf("paypal order error: [%s] %s", raw.Status, string(body))
	}

	if err := json.NewDecoder(raw.Body).Decode(&resp); err != nil {
		return resp, ErrRespondCapture
	}

	return resp, nil
}

func getAccessToken() (AccessToken, error) {
	at := AccessToken{}
	form := url.Values{}
	form.Set("grant_type", "client_credentials")

	req, err := http.NewRequest(
		http.MethodPost,
		config.PayPalEndpoint+"/"+tokenEndpoint,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return at, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(config.PayPalClientID, config.PayPalClientSecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return at, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return at, fmt.Errorf("token request failed: [%s] %s", resp.Status, body)
	}

	if err := json.NewDecoder(resp.Body).Decode(&at); err != nil {
		return at, err
	}

	return at, nil
}
