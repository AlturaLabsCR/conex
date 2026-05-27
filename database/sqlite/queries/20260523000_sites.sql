-- name: CreateSite :exec
INSERT INTO sites (sub, path, public)
VALUES (?, ?, ?);

-- name: SelectSiteByPath :one
SELECT sub, path, public
FROM sites
WHERE path = ?;

-- name: SelectSitesBySub :many
SELECT sub, path, public
FROM sites
WHERE sub = ?
ORDER BY path;

-- name: UpdateSitePublic :one
UPDATE sites
SET public = ?
WHERE sub = ? AND path = ?
RETURNING sub, path, public;

-- name: DeleteSite :one
DELETE FROM sites
WHERE sub = ? AND path = ?
RETURNING sub, path, public;
