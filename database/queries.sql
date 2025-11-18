-- name: GetUsers :many
SELECT * FROM users;

-- name: GetUserByID :one
SELECT * FROM users WHERE user_id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE user_email = $1;

-- name: GetSites :many
SELECT * FROM sites;

-- name: GetSlugs :many
SELECT site_slug FROM sites;

-- name: GetSiteByID :one
SELECT * FROM sites WHERE site_id = $1;

-- name: GetSiteBySlug :one
SELECT * FROM sites WHERE site_slug = $1;

-- name: GetActiveSites :many
SELECT * FROM sites WHERE site_deleted = 0;

-- name: GetPublishedSites :many
SELECT * FROM sites WHERE site_published = 1;

-- name: GetPublishedSitesWithMetrics :many
SELECT * FROM sites_with_metrics
ORDER BY metric_visits_total DESC,
site_id ASC LIMIT 30;

-- name: GetMetrics :many
SELECT * FROM site_metrics;

-- name: GetSyncData :one
SELECT * FROM site_sync WHERE site_sync_id = $1;

-- name: InsertSyncData :one
INSERT INTO site_sync(
  site_sync_id,
  site_sync_data_gz,
  site_sync_last_update_unix
) VALUES ($1, $2, $3) RETURNING site_sync_id;

-- name: UpdateSyncData :exec
UPDATE site_sync SET
  site_sync_data_gz = $1,
  site_sync_last_update_unix = $2
WHERE site_sync_id = $3;

-- name: GetMetricsBySiteID :one
SELECT * FROM site_metrics WHERE metric_site = $1;

-- name: GetSitesWithMetricsByUserID :many
SELECT * FROM sites_with_metrics WHERE site_user = $1;

-- name: GetSiteWithMetrics :one
SELECT * FROM sites_with_metrics WHERE site_slug = $1;

-- name: GetPublishedSiteWithMetricsBySlug :one
SELECT * FROM sites_with_metrics WHERE site_slug = $1 AND site_published = 1;

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
    metric_visits_total < $1
    OR (metric_visits_total = $2 AND site_id > $3)
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
    metric_visits_total > $1
    OR (metric_visits_total = $2 AND site_id > $3)
  )
ORDER BY metric_visits_total ASC, site_id ASC
LIMIT 30;

-- name: InsertUser :one
INSERT INTO users(
  user_email,
  user_created_unix, user_modified_unix,
  user_deleted
) VALUES ($1, $2, $3, $4) RETURNING user_id;

-- name: UpdateUser :exec
UPDATE users SET
  user_email = $1,
  user_modified_unix = $2
WHERE user_id = $3;

-- name: UserExists :one
SELECT EXISTS (
  SELECT 1
  FROM users
  WHERE user_id = $1
);

-- name: UserExistsEmail :one
SELECT EXISTS (
  SELECT 1
  FROM users
  WHERE user_email = $1
);

-- name: InsertSession :one
INSERT INTO sessions(
  "session_user",
  session_device,
  session_last_login_unix
) VALUES ($1, $2, $3) RETURNING session_id;

-- name: UpdateSession :exec
UPDATE sessions
SET session_last_login_unix = $1
WHERE session_id = $2;

-- name: SessionExists :one
SELECT EXISTS (
  SELECT 1
  FROM sessions
  WHERE session_id = $1
);

-- name: GetSession :one
SELECT * FROM sessions WHERE session_id = $1;

-- name: InsertPlan :one
INSERT INTO user_plans(
  user_plan_user,
  user_plan_created_unix,
  user_plan_modified_unix,
  user_plan_due_unix,
  user_plan_active
) VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdatePlan :exec
UPDATE user_plans SET
  user_plan_modified_unix = $1,
  user_plan_due_unix = $2,
  user_plan_active = $3
WHERE user_plan_id = $4;

-- name: GetPlan :one
SELECT * FROM user_plans WHERE user_plan_user = $1;

-- name: GetSessionsByUser :many
SELECT * FROM sessions WHERE "session_user" = $1
ORDER BY session_last_login_unix DESC;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE session_id = $1;

-- name: InsertSite :one
INSERT INTO sites (
  site_user,
  site_slug,
  site_title,
  site_tags_json,
  site_description,
  site_html_gz,
  site_created_unix,
  site_modified_unix,
  site_published,
  site_home_page,
  site_deleted
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING site_id;

-- name: UpdateSite :exec
UPDATE sites SET
  site_title = $1,
  site_description = $2,
  site_tags_json = $3,
  site_html_gz = $4,
  site_modified_unix = $5,
  site_published = $6,
  site_deleted = $7
WHERE site_id = $8;

-- name: PublishSite :exec
UPDATE sites SET
  site_published = 1
WHERE site_id = $1;

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
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateObject :exec
UPDATE site_objects SET
  object_mime = $1,
  object_md5 = $2,
  object_size_bytes = $3,
  object_modified_unix = $4
WHERE object_id = $5;

-- name: GetObjects :many
SELECT * FROM site_objects;

-- name: GetObject :one
SELECT * FROM site_objects WHERE object_bucket = $1 AND object_key = $2;

-- name: GetObjectByID :one
SELECT * FROM site_objects WHERE object_id = $1;

-- name: GetObjectsBySite :many
SELECT * FROM site_objects WHERE object_site = $1;

-- name: GetObjectByMD5 :one
SELECT * FROM site_objects WHERE object_md5 = $1;

-- name: UnpublishSite :exec
UPDATE sites SET
  site_published = 0
WHERE site_id = $1;

-- name: InsertMetric :one
INSERT INTO site_metrics (
  metric_site,
  metric_visits_total
) VALUES ($1, $2) RETURNING metric_id;

-- name: UpdateSiteSettings :exec
UPDATE sites SET
  site_modified_unix = $1,
  site_home_page = $2,
  site_tags_json = $3
WHERE site_id = $4;

-- name: DeleteObject :exec
DELETE FROM site_objects WHERE object_bucket = $1 AND object_key = $2;

-- name: GetBanner :one
SELECT * FROM site_banners WHERE banner_site = $1;

-- name: InsertBanner :one
INSERT INTO site_banners (
  banner_site,
  banner_object
) VALUES ($1, $2) RETURNING banner_id;

-- name: UpdateBanner :exec
UPDATE site_banners SET
  banner_object = $1
WHERE banner_id = $2;

-- name: DeleteBanner :exec
DELETE FROM site_banners WHERE banner_site = $1;

-- name: InsertPayment :one
INSERT INTO payments (
  payment_user,
  payment_amount,
  payment_date_unix,
  payment_successful,
  payment_reference
) VALUES ($1, $2, $3, $4, $5) RETURNING payment_id;

-- name: DeleteUser :exec
UPDATE users SET
  user_email = '',
  user_modified_unix = $1,
  user_deleted = 1
WHERE user_id = $2;

-- name: DeleteSite :exec
DELETE FROM sites WHERE site_id = $1;

-- name: NewVisit :exec
UPDATE site_metrics SET
metric_visits_total = metric_visits_total + 1
WHERE metric_site = $1;
