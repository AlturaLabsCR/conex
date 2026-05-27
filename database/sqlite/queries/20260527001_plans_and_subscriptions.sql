-- name: SelectAccountSubscriptionBySub :one
SELECT
  s.sub,
  p.id AS plan_id,
  p.name_key AS plan_name_key,
  p.price_amount,
  p.price_currency,
  p.billing_unit,
  p.billing_count,
  p.supports_renewal,
  s.status,
  s.due_date
FROM subscriptions s
JOIN plans p ON p.id = s.plan_id
WHERE s.sub = ?;
