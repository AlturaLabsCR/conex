package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"app/database"
	"app/middleware"
	"github.com/tavocg/go-paypal"
)

func (h *Handler) registerPlanRoutes() {
	authenticated := func(fn http.HandlerFunc) http.Handler {
		return middleware.AuthenticateBearer(h.logger, h.authenticator, h.localizeError, http.HandlerFunc(fn))
	}

	h.Add(http.MethodGet, h.routePath("/api/plans"), h.ListPlans)
	h.AddHandler(http.MethodPost, h.routePath("/api/plans/create/{planID}"), authenticated(h.CreatePlanOrder))
	h.AddHandler(http.MethodPost, h.routePath("/api/plans/capture/{orderID}"), authenticated(h.CapturePlanOrder))
}

func (h *Handler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.db.Querier().SelectPlans(r.Context())
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, err, "failed to list plans")
		return
	}

	L := h.localizer.LocalizerFunc(h.localizer.PickLanguageFromRequest(r))
	out := make([]planResponse, 0, len(plans))
	for _, plan := range plans {
		out = append(out, h.planResponse(L, plan))
	}

	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) CreatePlanOrder(w http.ResponseWriter, r *http.Request) {
	_, sub, status, ok := authenticatedAccountClaimsAndSubject(r)
	if !ok {
		h.writeStatus(w, r, status, "missing authenticated account subject")
		return
	}
	if h.paypal == nil {
		h.writeStatus(w, r, http.StatusServiceUnavailable, "paypal not configured", "sub", sub)
		return
	}

	plan, err := h.selectPlanByID(r.Context(), r.PathValue("planID"))
	if err != nil {
		if errors.Is(err, errPlanNotFound) {
			h.writeError(w, r, http.StatusNotFound, err, "plan not found", "sub", sub, "plan_id", r.PathValue("planID"))
			return
		}

		h.writeError(w, r, http.StatusInternalServerError, err, "failed to select plan", "sub", sub, "plan_id", r.PathValue("planID"))
		return
	}
	if plan.PriceAmount <= 0 {
		h.writeError(w, r, http.StatusBadRequest, errPlanNotPayable, "plan is not payable", "sub", sub, "plan_id", plan.ID)
		return
	}

	order, err := h.paypal.CreateOrder(
		r.Context(),
		plan.PriceCurrency,
		paypalAmount(plan.PriceAmount),
		paypal.WithoutOrderShipping(),
		paypal.WithOrderImmediatePayment(),
	)
	if err != nil {
		h.writeError(w, r, http.StatusBadGateway, err, "failed to create paypal order", "sub", sub, "plan_id", plan.ID)
		return
	}

	if err := h.db.Querier().CreatePayment(r.Context(), order.ID, sub, plan.ID, plan.PriceAmount, plan.PriceCurrency, "created"); err != nil {
		h.writeError(w, r, http.StatusInternalServerError, err, "failed to create payment", "sub", sub, "plan_id", plan.ID, "order_id", order.ID)
		return
	}

	writeJSON(w, http.StatusCreated, order)
}

func (h *Handler) CapturePlanOrder(w http.ResponseWriter, r *http.Request) {
	_, sub, status, ok := authenticatedAccountClaimsAndSubject(r)
	if !ok {
		h.writeStatus(w, r, status, "missing authenticated account subject")
		return
	}
	if h.paypal == nil {
		h.writeStatus(w, r, http.StatusServiceUnavailable, "paypal not configured", "sub", sub)
		return
	}

	orderID := r.PathValue("orderID")
	payment, err := h.db.Querier().SelectPaymentByOrderID(r.Context(), orderID)
	if err != nil {
		if h.db.IsErrNotFound(err) {
			h.writeError(w, r, http.StatusNotFound, err, "payment not found", "sub", sub, "order_id", orderID)
			return
		}

		h.writeError(w, r, http.StatusInternalServerError, err, "failed to select payment", "sub", sub, "order_id", orderID)
		return
	}
	if payment.Sub != sub {
		h.writeError(w, r, http.StatusNotFound, errPlanPaymentNotFound, "payment not found", "sub", sub, "order_id", orderID)
		return
	}

	capture, err := h.paypal.CaptureOrderPayment(r.Context(), orderID)
	if err != nil {
		h.writeError(w, r, http.StatusBadGateway, err, "failed to capture paypal order", "sub", sub, "order_id", orderID)
		return
	}
	if !paypalCaptureCompleted(capture) {
		h.writeError(w, r, http.StatusBadGateway, errPaypalCaptureIncomplete, "paypal capture incomplete", "sub", sub, "order_id", orderID)
		return
	}

	var captured *database.Payment
	dueDate := paymentDueDate(payment)
	if err := h.db.WithTx(r.Context(), func(q database.Querier) error {
		var err error
		captured, err = q.CapturePayment(r.Context(), orderID, sub)
		if err != nil {
			return err
		}

		return q.UpdateSubscriptionPlan(r.Context(), sub, captured.PlanID, dueDate)
	}); err != nil {
		h.writeError(w, r, http.StatusInternalServerError, err, "failed to capture payment", "sub", sub, "order_id", orderID)
		return
	}

	type captureResponse struct {
		Payment database.Payment                    `json:"payment"`
		PayPal  *paypal.CaptureOrderPaymentResponse `json:"paypal"`
	}

	writeJSON(w, http.StatusOK, captureResponse{
		Payment: *captured,
		PayPal:  capture,
	})
}

type priceResponse struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type billingPeriodResponse struct {
	Unit  string `json:"unit"`
	Count int64  `json:"count"`
}

type planPolicyResponse struct {
	MaxBytesPerSite    int64 `json:"max_bytes_per_site"`
	MaxSites           int64 `json:"max_sites"`
	MaxSubpathsPerSite int64 `json:"max_subpaths_per_site"`
}

type planResponse struct {
	ID              string                `json:"id"`
	Name            string                `json:"name"`
	Price           priceResponse         `json:"price"`
	BillingPeriod   billingPeriodResponse `json:"billing_period"`
	SupportsRenewal bool                  `json:"supports_renewal"`
	Policy          planPolicyResponse    `json:"policy"`
}

func (h *Handler) planResponse(L func(string, ...any) string, plan database.Plan) planResponse {
	return planResponse{
		ID:   plan.ID,
		Name: L(plan.NameKey),
		Price: priceResponse{
			Amount:   plan.PriceAmount,
			Currency: plan.PriceCurrency,
		},
		BillingPeriod: billingPeriodResponse{
			Unit:  plan.BillingUnit,
			Count: plan.BillingCount,
		},
		SupportsRenewal: plan.SupportsRenewal,
		Policy: planPolicyResponse{
			MaxBytesPerSite:    plan.Policy.MaxBytesPerSite,
			MaxSites:           plan.Policy.MaxSites,
			MaxSubpathsPerSite: plan.Policy.MaxSubpathsPerSite,
		},
	}
}

var (
	errPlanNotFound            = errors.New("plan not found")
	errPlanNotPayable          = errors.New("plan is not payable")
	errPlanPaymentNotFound     = errors.New("payment not found")
	errPaypalCaptureIncomplete = errors.New("paypal capture incomplete")
)

func (h *Handler) selectPlanByID(ctx context.Context, planID string) (database.Plan, error) {
	plans, err := h.db.Querier().SelectPlans(ctx)
	if err != nil {
		return database.Plan{}, err
	}

	for _, plan := range plans {
		if plan.ID == planID {
			return plan, nil
		}
	}

	return database.Plan{}, errPlanNotFound
}

func paypalAmount(cents int64) string {
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}

func paymentDueDate(payment *database.Payment) string {
	if payment == nil || payment.PlanID == "free" {
		return ""
	}

	return time.Now().UTC().AddDate(1, 0, 0).Format(time.DateOnly)
}

func paypalCaptureCompleted(capture *paypal.CaptureOrderPaymentResponse) bool {
	if capture == nil {
		return false
	}

	for _, unit := range capture.PurchaseUnits {
		for _, captured := range unit.Payments.Captures {
			if strings.EqualFold(captured.Status, "COMPLETED") {
				return true
			}
		}
	}

	return false
}
