CREATE TABLE IF NOT EXISTS plans (
  id TEXT PRIMARY KEY,
  name_key TEXT NOT NULL,
  price_amount BIGINT NOT NULL,
  price_currency TEXT NOT NULL,
  billing_unit TEXT NOT NULL,
  billing_count INTEGER NOT NULL,
  supports_renewal BOOLEAN NOT NULL
);

INSERT INTO plans (id, name_key, price_amount, price_currency, billing_unit, billing_count, supports_renewal)
VALUES ('free', 'plans.free.name', 0, 'USD', '', 0, FALSE)
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS subscriptions (
  sub BIGINT PRIMARY KEY REFERENCES accounts(sub) ON DELETE CASCADE,
  plan_id TEXT NOT NULL REFERENCES plans(id),
  status TEXT NOT NULL,
  due_date TEXT NOT NULL
);

INSERT INTO subscriptions (sub, plan_id, status, due_date)
SELECT a.sub, 'free', 'active', ''
FROM accounts a
ON CONFLICT (sub) DO NOTHING;

CREATE OR REPLACE FUNCTION create_default_subscription()
RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO subscriptions (sub, plan_id, status, due_date)
  VALUES (NEW.sub, 'free', 'active', '')
  ON CONFLICT (sub) DO NOTHING;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS accounts_default_subscription ON accounts;

CREATE TRIGGER accounts_default_subscription
AFTER INSERT ON accounts
FOR EACH ROW
EXECUTE FUNCTION create_default_subscription();
