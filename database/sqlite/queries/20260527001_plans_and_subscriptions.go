package queries

import (
	"context"

	"app/database"
)

func (q *SqliteQuerier) SelectAccountSubscriptionBySub(ctx context.Context, sub int64) (*database.AccountSubscription, error) {
	subscription, err := q.queries.SelectAccountSubscriptionBySub(ctx, sub)
	if err != nil {
		return nil, err
	}

	return &database.AccountSubscription{
		Sub:             subscription.Sub,
		PlanID:          subscription.PlanID,
		PlanNameKey:     subscription.PlanNameKey,
		PriceAmount:     subscription.PriceAmount,
		PriceCurrency:   subscription.PriceCurrency,
		BillingUnit:     subscription.BillingUnit,
		BillingCount:    subscription.BillingCount,
		SupportsRenewal: subscription.SupportsRenewal != 0,
		Status:          subscription.Status,
		DueDate:         subscription.DueDate,
	}, nil
}
