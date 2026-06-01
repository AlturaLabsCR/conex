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
