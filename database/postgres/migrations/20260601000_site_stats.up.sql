CREATE TABLE IF NOT EXISTS site_timestamps (
  path VARCHAR(255) NOT NULL PRIMARY KEY REFERENCES sites(path) ON DELETE CASCADE,
  created_at BIGINT NOT NULL,
  last_modified BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS site_clicks (
  path VARCHAR(255) NOT NULL PRIMARY KEY REFERENCES sites(path) ON DELETE CASCADE,
  clicks BIGINT NOT NULL DEFAULT 0,

  CONSTRAINT site_clicks_non_negative CHECK (clicks >= 0)
);

INSERT INTO site_timestamps (path, created_at, last_modified)
SELECT path, extract(epoch FROM now())::BIGINT, extract(epoch FROM now())::BIGINT
FROM sites
ON CONFLICT (path) DO NOTHING;

INSERT INTO site_clicks (path, clicks)
SELECT path, 0
FROM sites
ON CONFLICT (path) DO NOTHING;

CREATE OR REPLACE VIEW site_meta AS
SELECT
  s.sub,
  s.path,
  s.public,
  COALESCE(sn.name, s.path) AS name,
  COALESCE(
    (array_agg(st.tag ORDER BY st.tag) FILTER (WHERE st.tag IS NOT NULL))::text[],
    ARRAY[]::text[]
  ) AS tags,
  COALESCE(sts.created_at, 0) AS created_at,
  COALESCE(sts.last_modified, 0) AS last_modified,
  COALESCE(sc.clicks, 0) AS clicks
FROM sites s
LEFT JOIN site_names sn ON sn.path = s.path
LEFT JOIN site_tags st ON st.path = s.path
LEFT JOIN site_timestamps sts ON sts.path = s.path
LEFT JOIN site_clicks sc ON sc.path = s.path
GROUP BY s.sub, s.path, s.public, sn.name, sts.created_at, sts.last_modified, sc.clicks;
