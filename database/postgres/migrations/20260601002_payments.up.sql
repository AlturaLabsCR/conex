CREATE TABLE IF NOT EXISTS payments (
  order_id TEXT PRIMARY KEY,
  sub BIGINT NOT NULL REFERENCES accounts(sub) ON DELETE CASCADE,
  plan_id TEXT NOT NULL REFERENCES plans(id),
  amount BIGINT NOT NULL,
  currency TEXT NOT NULL,
  status TEXT NOT NULL,
  created_at BIGINT NOT NULL,
  paid_at BIGINT NOT NULL DEFAULT 0,

  CONSTRAINT payments_amount_non_negative CHECK (amount >= 0),
  CONSTRAINT payments_paid_at_non_negative CHECK (paid_at >= 0)
);

CREATE INDEX IF NOT EXISTS payments_sub_idx ON payments(sub);
CREATE INDEX IF NOT EXISTS payments_plan_idx ON payments(plan_id);
