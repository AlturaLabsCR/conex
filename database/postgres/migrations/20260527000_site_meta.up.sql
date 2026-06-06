CREATE TABLE IF NOT EXISTS site_names (
  path VARCHAR(255) NOT NULL PRIMARY KEY REFERENCES sites(path) ON DELETE CASCADE,
  name VARCHAR(255) NOT NULL,

  CONSTRAINT site_names_name_len CHECK (length(btrim(name)) BETWEEN 1 AND 255)
);

CREATE TABLE IF NOT EXISTS site_tags (
  path VARCHAR(255) NOT NULL REFERENCES sites(path) ON DELETE CASCADE,
  tag VARCHAR(16) NOT NULL,

  PRIMARY KEY (path, tag),
  CONSTRAINT site_tags_tag_len CHECK (length(btrim(tag)) BETWEEN 1 AND 16)
);

CREATE OR REPLACE VIEW site_meta AS
SELECT
  s.sub,
  s.path,
  s.public,
  COALESCE(sn.name, s.path) AS name,
  COALESCE(
    (array_agg(st.tag ORDER BY st.tag) FILTER (WHERE st.tag IS NOT NULL))::text[],
    ARRAY[]::text[]
  ) AS tags
FROM sites s
LEFT JOIN site_names sn ON sn.path = s.path
LEFT JOIN site_tags st ON st.path = s.path
GROUP BY s.sub, s.path, s.public, sn.name;
