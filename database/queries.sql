-- name: GetUsers :many
SELECT * FROM users;

-- name: GetUserByID :one
SELECT * FROM users WHERE user_id = ?;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE user_email = ?;

-- name: GetSites :many
SELECT * FROM sites;

-- name: GetSlugs :many
SELECT site_slug FROM sites;

-- name: GetSiteByID :one
SELECT * FROM sites WHERE site_id = ?;

-- name: GetSiteBySlug :one
SELECT * FROM sites WHERE site_slug = ?;

-- name: GetActiveSites :many
SELECT * FROM sites WHERE site_deleted = 0;

-- name: GetPublishedSites :many
SELECT * FROM sites WHERE site_published = 1;

-- name: GetMetrics :many
SELECT * FROM site_metrics;

-- name: GetMetricsBySiteID :one
SELECT * FROM site_metrics WHERE metric_site = ?;

-- name: GetPublishedSitesWithMetrics :many
SELECT * FROM sites_with_metrics
WHERE site_published = 1 AND site_deleted = 0;

-- name: GetSitesWithMetricsByUserID :many
SELECT * FROM sites_with_metrics WHERE site_user = ?;

-- name: GetSiteWithMetrics :one
SELECT * FROM sites_with_metrics WHERE site_slug = ?;

-- name: GetPublishedSiteWithMetricsBySlug :one
SELECT * FROM sites_with_metrics WHERE site_slug = ?;

-- name: GetPublishedSitesWithMetricsFromMostTotalVisits :many
SELECT * FROM sites_with_metrics
WHERE site_published = 1 AND site_deleted = 0
ORDER BY metric_visits_total DESC, site_id ASC
LIMIT 30;

-- name: GetValidSitesWithMetricsFromMostTotalVisitsLessThan :many
SELECT *
FROM sites_with_metrics
WHERE site_published = 1
  AND site_deleted = 0
  AND (
    metric_visits_total < ?
    OR (metric_visits_total = ? AND site_id > ?)
  )
ORDER BY metric_visits_total DESC, site_id ASC
LIMIT 30;

-- name: GetValidSitesWithMetricsFromLeastTotalVisits :many
SELECT *
FROM sites_with_metrics
WHERE site_published = 1 AND site_deleted = 0
ORDER BY metric_visits_total ASC, site_id ASC
LIMIT 30;

-- name: GetValidSitesWithMetricsFromLeastTotalVisitsMoreThan :many
SELECT *
FROM sites_with_metrics
WHERE site_published = 1
  AND site_deleted = 0
  AND (
    metric_visits_total > ?
    OR (metric_visits_total = ? AND site_id > ?)
  )
ORDER BY metric_visits_total ASC, site_id ASC
LIMIT 30;

-- name: InsertUser :one
INSERT INTO users(
  user_email, user_name,
  user_created_unix, user_modified_unix,
  user_deleted
) VALUES (?, ?, ?, ?, ?) RETURNING user_id;

-- name: UserExists :one
SELECT EXISTS (
  SELECT 1
  FROM users
  WHERE user_id = ?
);

-- name: UserExistsEmail :one
SELECT EXISTS (
  SELECT 1
  FROM users
  WHERE user_email = ?
);

-- name: InsertSession :one
INSERT INTO sessions(
  session_user,
  session_device,
  session_last_login_unix
) VALUES (?, ?, ?) RETURNING session_id;

-- name: UpdateSession :exec
UPDATE sessions
SET session_last_login_unix = ?
WHERE session_id = ?;

-- name: SessionExists :one
SELECT EXISTS (
  SELECT 1
  FROM sessions
  WHERE session_id = ?
);

-- name: GetSession :one
SELECT * FROM sessions WHERE session_id = ?;

-- name: GetSessionsByUser :many
SELECT * FROM sessions WHERE session_user = ?
ORDER BY session_last_login_unix DESC;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE session_id = ?;

-- name: InsertSite :one
INSERT INTO sites (
  site_user,
  site_slug,
  site_title,
  site_tags_json,
  site_description,
  site_html_published,
  site_created_unix,
  site_modified_unix,
  site_published,
  site_deleted
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING site_id;

-- name: UpdateSite :exec
UPDATE sites SET
  site_title = ?,
  site_description = ?,
  site_tags_json = ?,
  site_html_published = ?,
  site_modified_unix = ?,
  site_published = ?,
  site_deleted = ?
WHERE site_id = ?;

-- name: PublishSite :exec
UPDATE sites SET
  site_published = 1
WHERE site_id = ?;

-- name: UnpublishSite :exec
UPDATE sites SET
  site_published = 0
WHERE site_id = ?;

-- name: InsertMetric :one
INSERT INTO site_metrics (
  metric_site,
  metric_visits_total
) VALUES (?, ?) RETURNING metric_id;

-- name: UpdateTags :exec
UPDATE sites SET
  site_tags_json = ?
WHERE site_id = ?;
