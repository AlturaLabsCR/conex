package queries

import (
	"context"

	"app/database"
	"app/database/sqlite/db"
)

func (q *SqliteQuerier) CreatePayment(ctx context.Context, orderID string, sub int64, planID string, amount int64, currency string, status string) error {
	return q.queries.CreatePayment(ctx, db.CreatePaymentParams{
		OrderID:  orderID,
		Sub:      sub,
		PlanID:   planID,
		Amount:   amount,
		Currency: currency,
		Status:   status,
	})
}

func (q *SqliteQuerier) SelectPaymentByOrderID(ctx context.Context, orderID string) (*database.Payment, error) {
	payment, err := q.queries.SelectPaymentByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	return databasePayment(payment), nil
}

func (q *SqliteQuerier) CapturePayment(ctx context.Context, orderID string, sub int64) (*database.Payment, error) {
	payment, err := q.queries.CapturePayment(ctx, db.CapturePaymentParams{
		OrderID: orderID,
		Sub:     sub,
	})
	if err != nil {
		return nil, err
	}

	return databasePayment(payment), nil
}

func (q *SqliteQuerier) UpdateSubscriptionPlan(ctx context.Context, sub int64, planID string, dueDate string) error {
	return q.queries.UpdateSubscriptionPlan(ctx, db.UpdateSubscriptionPlanParams{
		PlanID:  planID,
		DueDate: dueDate,
		Sub:     sub,
	})
}

func databasePayment(payment db.Payment) *database.Payment {
	return &database.Payment{
		OrderID:   payment.OrderID,
		Sub:       payment.Sub,
		PlanID:    payment.PlanID,
		Amount:    payment.Amount,
		Currency:  payment.Currency,
		Status:    payment.Status,
		CreatedAt: payment.CreatedAt,
		PaidAt:    payment.PaidAt,
	}
}
