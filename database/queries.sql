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

-- name: GetSyncData :one
SELECT * FROM site_sync WHERE site_sync_id = ?;

-- name: InsertSyncData :one
INSERT INTO site_sync(
  site_sync_id,
  site_sync_data_staging,
  site_sync_last_update_unix
) VALUES (?, ?, ?) RETURNING site_sync_id;

-- name: UpdateSyncData :exec
UPDATE site_sync SET
  site_sync_data_staging = ?,
  site_sync_last_update_unix = ?
WHERE site_sync_id = ?;

-- name: GetMetricsBySiteID :one
SELECT * FROM site_metrics WHERE metric_site = ?;

-- name: GetSitesWithMetricsByUserID :many
SELECT * FROM sites_with_metrics WHERE site_user = ?;

-- name: GetSiteWithMetrics :one
SELECT * FROM sites_with_metrics WHERE site_slug = ?;

-- name: GetPublishedSiteWithMetricsBySlug :one
SELECT * FROM sites_with_metrics WHERE site_slug = ? AND site_published = 1;

-- name: GetHomePageSitesWithMetricsFromMostTotalVisits :many
SELECT * FROM sites_with_metrics
WHERE site_published = 1 AND site_home_page = 1 AND site_deleted = 0
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
  user_email,
  user_created_unix, user_modified_unix,
  user_deleted
) VALUES (?, ?, ?, ?) RETURNING user_id;

-- name: UpdateUser :exec
UPDATE users SET
  user_email = ?,
  user_modified_unix = ?
WHERE user_id = ?;

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

-- name: InsertPlan :one
INSERT INTO user_plans(
  user_plan_user,
  user_plan_created_unix,
  user_plan_modified_unix,
  user_plan_due_unix,
  user_plan_active
) VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdatePlan :exec
UPDATE user_plans SET
  user_plan_modified_unix = ?,
  user_plan_due_unix = ?,
  user_plan_active = ?
WHERE user_plan_id = ?;

-- name: GetPlan :one
SELECT * FROM user_plans WHERE user_plan_user = ?;

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
  site_home_page,
  site_deleted
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
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

-- name: InsertObject :one
INSERT INTO site_objects(
  object_site,
  object_bucket,
  object_key,
  object_mime,
  object_md5,
  object_size_bytes,
  object_created_unix,
  object_modified_unix
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateObject :exec
UPDATE site_objects SET
  object_mime = ?,
  object_md5 = ?,
  object_size_bytes = ?,
  object_modified_unix = ?
WHERE object_id = ?;

-- name: GetObjects :many
SELECT * FROM site_objects;

-- name: GetObject :one
SELECT * FROM site_objects WHERE object_bucket = ? AND object_key = ?;

-- name: GetObjectByID :one
SELECT * FROM site_objects WHERE object_id = ?;

-- name: GetObjectsBySite :many
SELECT * FROM site_objects WHERE object_site = ?;

-- name: GetObjectByMD5 :one
SELECT * FROM site_objects WHERE object_md5 = ?;

-- name: UnpublishSite :exec
UPDATE sites SET
  site_published = 0
WHERE site_id = ?;

-- name: InsertMetric :one
INSERT INTO site_metrics (
  metric_site,
  metric_visits_total
) VALUES (?, ?) RETURNING metric_id;

-- name: UpdateSiteSettings :exec
UPDATE sites SET
  site_modified_unix = ?,
  site_home_page = ?,
  site_tags_json = ?
WHERE site_id = ?;

-- name: DeleteObject :exec
DELETE FROM site_objects WHERE object_bucket = ? AND object_key = ?;

-- name: GetBanner :one
SELECT * FROM site_banners WHERE banner_site = ?;

-- name: InsertBanner :one
INSERT INTO site_banners (
  banner_site,
  banner_object
) VALUES (?, ?) RETURNING banner_id;

-- name: UpdateBanner :exec
UPDATE site_banners SET
  banner_object = ?
WHERE banner_id = ?;

-- name: DeleteBanner :exec
DELETE FROM site_banners WHERE banner_site = ?;

-- name: InsertPayment :one
INSERT INTO payments (
  payment_user,
  payment_amount,
  payment_date_unix,
  payment_successful,
  payment_reference
) VALUES (?, ?, ?, ?, ?) RETURNING payment_id;

-- name: DeleteUser :exec
UPDATE users SET
  user_email = '',
  user_modified_unix = ?,
  user_deleted = 1
WHERE user_id = ?;

-- name: DeleteSite :exec
DELETE FROM sites WHERE site_id = ?;
