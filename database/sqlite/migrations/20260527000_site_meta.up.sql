CREATE TABLE IF NOT EXISTS site_names (
  path VARCHAR(255) NOT NULL PRIMARY KEY REFERENCES sites(path) ON DELETE CASCADE,
  name VARCHAR(255) NOT NULL,

  CONSTRAINT site_names_name_len CHECK (length(trim(name)) BETWEEN 1 AND 255)
);

CREATE TABLE IF NOT EXISTS site_tags (
  path VARCHAR(255) NOT NULL REFERENCES sites(path) ON DELETE CASCADE,
  tag VARCHAR(16) NOT NULL,

  PRIMARY KEY (path, tag),
  CONSTRAINT site_tags_tag_len CHECK (length(trim(tag)) BETWEEN 1 AND 16)
);

CREATE VIEW IF NOT EXISTS site_meta AS
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
  ) AS tags
FROM sites s
LEFT JOIN site_names sn ON sn.path = s.path;
