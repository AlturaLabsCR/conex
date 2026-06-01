CREATE TABLE IF NOT EXISTS plan_policies (
  plan_id TEXT NOT NULL PRIMARY KEY REFERENCES plans(id) ON DELETE CASCADE,
  policy TEXT NOT NULL CHECK (json_valid(policy))
);

INSERT INTO plans (id, name_key, price_amount, price_currency, billing_unit, billing_count, supports_renewal)
VALUES
  ('free', 'plans.free.name', 0, 'USD', '', 0, 0),
  ('creator', 'plans.creator.name', 2499, 'USD', 'year', 1, 1),
  ('pro', 'plans.pro.name', 7999, 'USD', 'year', 1, 1)
ON CONFLICT (id) DO UPDATE
SET
  name_key = excluded.name_key,
  price_amount = excluded.price_amount,
  price_currency = excluded.price_currency,
  billing_unit = excluded.billing_unit,
  billing_count = excluded.billing_count,
  supports_renewal = excluded.supports_renewal;

INSERT INTO plan_policies (plan_id, policy)
VALUES
  ('free', '{"max_bytes_per_site":1048576,"max_sites":1,"max_subpaths_per_site":0}'),
  ('creator', '{"max_bytes_per_site":10485760,"max_sites":3,"max_subpaths_per_site":2}'),
  ('pro', '{"max_bytes_per_site":104857600,"max_sites":50,"max_subpaths_per_site":10}')
ON CONFLICT (plan_id) DO UPDATE
SET policy = excluded.policy;

DROP VIEW IF EXISTS plan_meta;

CREATE VIEW plan_meta AS
SELECT
  p.id,
  p.name_key,
  p.price_amount,
  p.price_currency,
  p.billing_unit,
  p.billing_count,
  p.supports_renewal,
  pp.policy
FROM plans p
LEFT JOIN plan_policies pp ON pp.plan_id = p.id;
