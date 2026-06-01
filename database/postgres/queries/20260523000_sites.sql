-- name: CreateSite :exec
INSERT INTO sites (sub, path, public)
VALUES ($1, $2, $3);

-- name: CreateSiteTimestamps :exec
INSERT INTO site_timestamps (path, created_at, last_modified)
VALUES ($1, extract(epoch FROM now())::BIGINT, extract(epoch FROM now())::BIGINT);

-- name: CreateSiteClicks :exec
INSERT INTO site_clicks (path, clicks)
VALUES ($1, 0);

-- name: UpsertSiteName :exec
INSERT INTO site_names (path, name)
VALUES ($1, $2)
ON CONFLICT (path) DO UPDATE
SET name = EXCLUDED.name;

-- name: DeleteSiteTags :exec
DELETE FROM site_tags
WHERE path = $1;

-- name: CreateSiteTag :exec
INSERT INTO site_tags (path, tag)
VALUES ($1, $2);

-- name: SelectSiteByPath :one
SELECT sub, path, public, name, to_json(tags)::text AS tags, created_at, last_modified, clicks
FROM site_meta
WHERE path = $1;

-- name: SelectSitesBySub :many
SELECT sub, path, public, name, to_json(tags)::text AS tags, created_at, last_modified, clicks
FROM site_meta
WHERE sub = $1
ORDER BY path;

-- name: SelectPublicSitesByClicks :many
SELECT sub, path, public, name, to_json(tags)::text AS tags, created_at, last_modified, clicks
FROM site_meta
WHERE public = TRUE
ORDER BY clicks DESC, path
LIMIT $1 OFFSET $2;

-- name: SelectPublicSitesByCreatedAt :many
SELECT sub, path, public, name, to_json(tags)::text AS tags, created_at, last_modified, clicks
FROM site_meta
WHERE public = TRUE
ORDER BY created_at DESC, path
LIMIT $1 OFFSET $2;

-- name: SearchPublicSites :many
WITH search AS (
  SELECT lower(btrim(sqlc.arg(query)::text)) AS query,
    sqlc.arg(page_limit)::int AS limit_value,
    sqlc.arg(page_offset)::int AS offset_value
)
SELECT sub, path, public, name, to_json(tags)::text AS tags, created_at, last_modified, clicks
FROM site_meta, search
WHERE public = TRUE
  AND (
    lower(name) LIKE '%' || search.query || '%'
    OR EXISTS (
      SELECT 1
      FROM site_tags
      WHERE site_tags.path = site_meta.path
        AND lower(site_tags.tag) LIKE '%' || search.query || '%'
    )
  )
ORDER BY
  CASE
    WHEN lower(name) = search.query THEN 0
    WHEN EXISTS (
      SELECT 1
      FROM site_tags
      WHERE site_tags.path = site_meta.path
        AND lower(site_tags.tag) = search.query
    ) THEN 1
    WHEN lower(name) LIKE search.query || '%' THEN 2
    WHEN EXISTS (
      SELECT 1
      FROM site_tags
      WHERE site_tags.path = site_meta.path
        AND lower(site_tags.tag) LIKE search.query || '%'
    ) THEN 3
    WHEN lower(name) LIKE '%' || search.query || '%' THEN 4
    ELSE 5
  END,
  clicks DESC,
  created_at DESC,
  path
LIMIT (SELECT limit_value FROM search)
OFFSET (SELECT offset_value FROM search);

-- name: UpdateSitePublic :exec
UPDATE sites
SET public = $3
WHERE sub = $1 AND path = $2;

-- name: IncrementSiteClicks :exec
UPDATE site_clicks
SET clicks = clicks + 1
WHERE path = $1;

-- name: UpdateSiteLastModified :exec
UPDATE site_timestamps
SET last_modified = extract(epoch FROM now())::BIGINT
WHERE path = $1;

-- name: DeleteSite :one
DELETE FROM sites
WHERE sub = $1 AND path = $2
RETURNING sub, path, public, path AS name, '[]' AS tags, 0::BIGINT AS created_at, 0::BIGINT AS last_modified, 0::BIGINT AS clicks;
