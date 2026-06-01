-- name: CreatePayment :exec
INSERT INTO payments (order_id, sub, plan_id, amount, currency, status, created_at, paid_at)
VALUES (?, ?, ?, ?, ?, ?, unixepoch(), 0);

-- name: SelectPaymentByOrderID :one
SELECT order_id, sub, plan_id, amount, currency, status, created_at, paid_at
FROM payments
WHERE order_id = ?;

-- name: CapturePayment :one
UPDATE payments
SET status = 'captured',
    paid_at = unixepoch()
WHERE order_id = ? AND sub = ?
RETURNING order_id, sub, plan_id, amount, currency, status, created_at, paid_at;

-- name: UpdateSubscriptionPlan :exec
UPDATE subscriptions
SET plan_id = ?,
    status = 'active',
    due_date = ?
WHERE sub = ?;
