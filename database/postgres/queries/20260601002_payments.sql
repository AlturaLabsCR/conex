-- name: CreatePayment :exec
INSERT INTO payments (order_id, sub, plan_id, amount, currency, status, created_at, paid_at)
VALUES ($1, $2, $3, $4, $5, $6, extract(epoch FROM now())::BIGINT, 0);

-- name: SelectPaymentByOrderID :one
SELECT order_id, sub, plan_id, amount, currency, status, created_at, paid_at
FROM payments
WHERE order_id = $1;

-- name: CapturePayment :one
UPDATE payments
SET status = 'captured',
    paid_at = extract(epoch FROM now())::BIGINT
WHERE order_id = $1 AND sub = $2
RETURNING order_id, sub, plan_id, amount, currency, status, created_at, paid_at;

-- name: UpdateSubscriptionPlan :exec
UPDATE subscriptions
SET plan_id = $2,
    status = 'active',
    due_date = $3
WHERE sub = $1;
