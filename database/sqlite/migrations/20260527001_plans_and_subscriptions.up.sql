CREATE TABLE IF NOT EXISTS plans (
  id TEXT PRIMARY KEY,
  name_key TEXT NOT NULL,
  price_amount INTEGER NOT NULL,
  price_currency TEXT NOT NULL,
  billing_unit TEXT NOT NULL,
  billing_count INTEGER NOT NULL,
  supports_renewal INTEGER NOT NULL
);

INSERT OR IGNORE INTO plans (id, name_key, price_amount, price_currency, billing_unit, billing_count, supports_renewal)
VALUES ('free', 'plans.free.name', 0, 'USD', '', 0, 0);

CREATE TABLE IF NOT EXISTS subscriptions (
  sub INTEGER PRIMARY KEY REFERENCES accounts(sub) ON DELETE CASCADE,
  plan_id TEXT NOT NULL REFERENCES plans(id),
  status TEXT NOT NULL,
  due_date TEXT NOT NULL
);

INSERT OR IGNORE INTO subscriptions (sub, plan_id, status, due_date)
SELECT sub, 'free', 'active', ''
FROM accounts;

CREATE TRIGGER IF NOT EXISTS accounts_default_subscription
AFTER INSERT ON accounts
FOR EACH ROW
BEGIN
  INSERT OR IGNORE INTO subscriptions (sub, plan_id, status, due_date)
  VALUES (NEW.sub, 'free', 'active', '');
END;
