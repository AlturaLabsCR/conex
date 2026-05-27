-- name: CreateSite :exec
INSERT INTO sites (sub, path, public)
VALUES ($1, $2, $3);

-- name: SelectSiteByPath :one
SELECT sub, path, public
FROM sites
WHERE path = $1;

-- name: SelectSitesBySub :many
SELECT sub, path, public
FROM sites
WHERE sub = $1
ORDER BY path;

-- name: UpdateSitePublic :one
UPDATE sites
SET public = $3
WHERE sub = $1 AND path = $2
RETURNING sub, path, public;

-- name: DeleteSite :one
DELETE FROM sites
WHERE sub = $1 AND path = $2
RETURNING sub, path, public;
