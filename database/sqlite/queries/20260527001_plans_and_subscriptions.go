package queries

import (
	"context"
	"encoding/json"

	"app/database"
)

func (q *SqliteQuerier) SelectAccountSubscriptionBySub(ctx context.Context, sub int64) (*database.AccountSubscription, error) {
	subscription, err := q.queries.SelectAccountSubscriptionBySub(ctx, sub)
	if err != nil {
		return nil, err
	}
	var policy database.PlanPolicy
	if err := json.Unmarshal([]byte(subscription.Policy), &policy); err != nil {
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
		Policy:          policy,
		Status:          subscription.Status,
		DueDate:         subscription.DueDate,
	}, nil
}

func (q *SqliteQuerier) SelectPlans(ctx context.Context) ([]database.Plan, error) {
	plans, err := q.queries.SelectPlans(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]database.Plan, 0, len(plans))
	for _, plan := range plans {
		var policy database.PlanPolicy
		if err := json.Unmarshal([]byte(plan.Policy), &policy); err != nil {
			return nil, err
		}
		out = append(out, database.Plan{
			ID:              plan.ID,
			NameKey:         plan.NameKey,
			PriceAmount:     plan.PriceAmount,
			PriceCurrency:   plan.PriceCurrency,
			BillingUnit:     plan.BillingUnit,
			BillingCount:    plan.BillingCount,
			SupportsRenewal: plan.SupportsRenewal != 0,
			Policy:          policy,
		})
	}

	return out, nil
}
