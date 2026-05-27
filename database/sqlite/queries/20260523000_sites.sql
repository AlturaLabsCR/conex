-- name: CreateSite :exec
INSERT INTO sites (sub, path, public)
VALUES (?, ?, ?);

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
SELECT sub, path, public, name, CAST(tags AS TEXT) AS tags
FROM site_meta
WHERE path = ?;

-- name: SelectSitesBySub :many
SELECT sub, path, public, name, CAST(tags AS TEXT) AS tags
FROM site_meta
WHERE sub = ?
ORDER BY path;

-- name: UpdateSitePublic :exec
UPDATE sites
SET public = ?
WHERE sub = ? AND path = ?;

-- name: DeleteSite :one
DELETE FROM sites
WHERE sub = ? AND path = ?
RETURNING sub, path, public, path AS name, '[]' AS tags;
