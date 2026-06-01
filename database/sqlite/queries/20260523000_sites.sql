-- name: CreateSite :exec
INSERT INTO sites (sub, path, public)
VALUES (?, ?, ?);

-- name: CreateSiteTimestamps :exec
INSERT INTO site_timestamps (path, created_at, last_modified)
VALUES (?, unixepoch(), unixepoch());

-- name: CreateSiteClicks :exec
INSERT INTO site_clicks (path, clicks)
VALUES (?, 0);

-- name: UpsertSiteName :exec
INSERT INTO site_names (path, name)
VALUES (?, ?)
ON CONFLICT (path) DO UPDATE
SET name = excluded.name;

-- name: DeleteSiteTags :exec
DELETE FROM site_tags
WHERE path = ?;

-- name: CreateSiteTag :exec
INSERT INTO site_tags (path, tag)
VALUES (?, ?);

-- name: SelectSiteByPath :one
SELECT sub, path, public, name, CAST(tags AS TEXT) AS tags, created_at, last_modified, clicks
FROM site_meta
WHERE path = ?;

-- name: SelectSitesBySub :many
SELECT sub, path, public, name, CAST(tags AS TEXT) AS tags, created_at, last_modified, clicks
FROM site_meta
WHERE sub = ?
ORDER BY path;

-- name: SelectPublicSitesByClicks :many
SELECT sub, path, public, name, CAST(tags AS TEXT) AS tags, created_at, last_modified, clicks
FROM site_meta
WHERE public = TRUE
ORDER BY clicks DESC, path
LIMIT ? OFFSET ?;

-- name: SelectPublicSitesByCreatedAt :many
SELECT sub, path, public, name, CAST(tags AS TEXT) AS tags, created_at, last_modified, clicks
FROM site_meta
WHERE public = TRUE
ORDER BY created_at DESC, path
LIMIT ? OFFSET ?;

-- name: SearchPublicSites :many
SELECT sub, path, public, name, CAST(tags AS TEXT) AS tags, created_at, last_modified, clicks
FROM site_meta
WHERE public = TRUE
  AND (
    lower(name) LIKE '%' || lower(trim(CAST(?1 AS TEXT))) || '%'
    OR EXISTS (
      SELECT 1
      FROM site_tags
      WHERE site_tags.path = site_meta.path
        AND lower(site_tags.tag) LIKE '%' || lower(trim(CAST(?1 AS TEXT))) || '%'
    )
  )
ORDER BY
  CASE
    WHEN lower(name) = lower(trim(CAST(?1 AS TEXT))) THEN 0
    WHEN EXISTS (
      SELECT 1
      FROM site_tags
      WHERE site_tags.path = site_meta.path
        AND lower(site_tags.tag) = lower(trim(CAST(?1 AS TEXT)))
    ) THEN 1
    WHEN lower(name) LIKE lower(trim(CAST(?1 AS TEXT))) || '%' THEN 2
    WHEN EXISTS (
      SELECT 1
      FROM site_tags
      WHERE site_tags.path = site_meta.path
        AND lower(site_tags.tag) LIKE lower(trim(CAST(?1 AS TEXT))) || '%'
    ) THEN 3
    WHEN lower(name) LIKE '%' || lower(trim(CAST(?1 AS TEXT))) || '%' THEN 4
    ELSE 5
  END,
  clicks DESC,
  created_at DESC,
  path
LIMIT CAST(?2 AS INTEGER) OFFSET CAST(?3 AS INTEGER);

-- name: UpdateSitePublic :exec
UPDATE sites
SET public = ?
WHERE sub = ? AND path = ?;

-- name: IncrementSiteClicks :exec
UPDATE site_clicks
SET clicks = clicks + 1
WHERE path = ?;

-- name: UpdateSiteLastModified :exec
UPDATE site_timestamps
SET last_modified = unixepoch()
WHERE path = ?;

-- name: DeleteSite :one
DELETE FROM sites
WHERE sub = ? AND path = ?
RETURNING sub, path, public, path AS name, '[]' AS tags, 0 AS created_at, 0 AS last_modified, 0 AS clicks;
