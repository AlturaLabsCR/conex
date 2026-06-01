CREATE TABLE IF NOT EXISTS plan_policies (
  plan_id TEXT NOT NULL PRIMARY KEY REFERENCES plans(id) ON DELETE CASCADE,
  policy JSONB NOT NULL
);

INSERT INTO plans (id, name_key, price_amount, price_currency, billing_unit, billing_count, supports_renewal)
VALUES
  ('free', 'plans.free.name', 0, 'USD', '', 0, FALSE),
  ('creator', 'plans.creator.name', 2499, 'USD', 'year', 1, TRUE),
  ('pro', 'plans.pro.name', 7999, 'USD', 'year', 1, TRUE)
ON CONFLICT (id) DO UPDATE
SET
  name_key = EXCLUDED.name_key,
  price_amount = EXCLUDED.price_amount,
  price_currency = EXCLUDED.price_currency,
  billing_unit = EXCLUDED.billing_unit,
  billing_count = EXCLUDED.billing_count,
  supports_renewal = EXCLUDED.supports_renewal;

INSERT INTO plan_policies (plan_id, policy)
VALUES
  ('free', '{"max_bytes_per_site":1048576,"max_sites":1,"max_subpaths_per_site":0}'::jsonb),
  ('creator', '{"max_bytes_per_site":10485760,"max_sites":3,"max_subpaths_per_site":2}'::jsonb),
  ('pro', '{"max_bytes_per_site":104857600,"max_sites":50,"max_subpaths_per_site":10}'::jsonb)
ON CONFLICT (plan_id) DO UPDATE
SET policy = EXCLUDED.policy;

CREATE OR REPLACE VIEW plan_meta AS
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
