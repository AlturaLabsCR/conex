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
  COALESCE(p.policy, '{}') AS policy,
  s.status,
  s.due_date
FROM subscriptions s
JOIN plan_meta p ON p.id = s.plan_id
WHERE s.sub = ?;

-- name: SelectPlans :many
SELECT
  id,
  name_key,
  price_amount,
  price_currency,
  billing_unit,
  billing_count,
  supports_renewal,
  COALESCE(policy, '{}') AS policy
FROM plan_meta
ORDER BY price_amount, id;
