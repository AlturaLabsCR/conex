CREATE TABLE IF NOT EXISTS site_timestamps (
  path VARCHAR(255) NOT NULL PRIMARY KEY REFERENCES sites(path) ON DELETE CASCADE,
  created_at INTEGER NOT NULL,
  last_modified INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS site_clicks (
  path VARCHAR(255) NOT NULL PRIMARY KEY REFERENCES sites(path) ON DELETE CASCADE,
  clicks INTEGER NOT NULL DEFAULT 0,

  CONSTRAINT site_clicks_non_negative CHECK (clicks >= 0)
);

INSERT OR IGNORE INTO site_timestamps (path, created_at, last_modified)
SELECT path, unixepoch(), unixepoch()
FROM sites;

INSERT OR IGNORE INTO site_clicks (path, clicks)
SELECT path, 0
FROM sites;

DROP VIEW IF EXISTS site_meta;

CREATE VIEW site_meta AS
SELECT
  s.sub,
  s.path,
  s.public,
  COALESCE(sn.name, s.path) AS name,
  COALESCE(
    (
      SELECT json_group_array(ordered_tags.tag)
      FROM (
        SELECT tag
        FROM site_tags
        WHERE site_tags.path = s.path
        ORDER BY tag
      ) AS ordered_tags
    ),
    '[]'
  ) AS tags,
  COALESCE(sts.created_at, 0) AS created_at,
  COALESCE(sts.last_modified, 0) AS last_modified,
  COALESCE(sc.clicks, 0) AS clicks
FROM sites s
LEFT JOIN site_names sn ON sn.path = s.path
LEFT JOIN site_timestamps sts ON sts.path = s.path
LEFT JOIN site_clicks sc ON sc.path = s.path;
