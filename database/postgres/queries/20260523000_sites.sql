-- name: CreateSite :exec
INSERT INTO sites (sub, path, public)
VALUES ($1, $2, $3);

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
SELECT sub, path, public, name, to_json(tags)::text AS tags
FROM site_meta
WHERE path = $1;

-- name: SelectSitesBySub :many
SELECT sub, path, public, name, to_json(tags)::text AS tags
FROM site_meta
WHERE sub = $1
ORDER BY path;

-- name: UpdateSitePublic :exec
UPDATE sites
SET public = $3
WHERE sub = $1 AND path = $2;

-- name: DeleteSite :one
DELETE FROM sites
WHERE sub = $1 AND path = $2
RETURNING sub, path, public, path AS name, '[]' AS tags;
