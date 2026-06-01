package queries

import (
	"context"
	"encoding/json"
	"fmt"

	"app/database"
)

func (q *PostgresQuerier) SelectAccountSubscriptionBySub(ctx context.Context, sub int64) (*database.AccountSubscription, error) {
	subscription, err := q.queries.SelectAccountSubscriptionBySub(ctx, sub)
	if err != nil {
		return nil, err
	}
	policy, err := databasePlanPolicy(subscription.Policy)
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
		BillingCount:    int64(subscription.BillingCount),
		SupportsRenewal: subscription.SupportsRenewal,
		Policy:          policy,
		Status:          subscription.Status,
		DueDate:         subscription.DueDate,
	}, nil
}

func (q *PostgresQuerier) SelectPlans(ctx context.Context) ([]database.Plan, error) {
	plans, err := q.queries.SelectPlans(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]database.Plan, 0, len(plans))
	for _, plan := range plans {
		policy, err := databasePlanPolicy(plan.Policy)
		if err != nil {
			return nil, err
		}
		out = append(out, database.Plan{
			ID:              plan.ID,
			NameKey:         plan.NameKey,
			PriceAmount:     plan.PriceAmount,
			PriceCurrency:   plan.PriceCurrency,
			BillingUnit:     plan.BillingUnit,
			BillingCount:    int64(plan.BillingCount),
			SupportsRenewal: plan.SupportsRenewal,
			Policy:          policy,
		})
	}

	return out, nil
}

func databasePlanPolicy(policy any) (database.PlanPolicy, error) {
	var out database.PlanPolicy
	switch policy := policy.(type) {
	case nil:
		return out, nil
	case string:
		return out, json.Unmarshal([]byte(policy), &out)
	case []byte:
		return out, json.Unmarshal(policy, &out)
	default:
		return out, json.Unmarshal([]byte(fmt.Sprint(policy)), &out)
	}
}
