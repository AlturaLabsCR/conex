-- name: GetUsers :many
SELECT * FROM users;

-- name: GetUserByID :one
SELECT * FROM users WHERE user_id = ?;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE user_email = ?;

-- name: GetSites :many
SELECT * FROM sites;

-- name: GetSiteByID :one
SELECT * FROM sites WHERE site_id = ?;

-- name: GetSiteBySlug :one
SELECT * FROM sites WHERE site_slug = ?;

-- name: GetSitesByUserID :many
SELECT * FROM sites WHERE site_user = ?;

-- name: GetActiveSites :many
SELECT * FROM sites WHERE site_deleted = 0;

-- name: GetPublishedSites :many
SELECT * FROM sites WHERE site_published = 1;

-- name: GetTags :many
SELECT * FROM site_tags;

-- name: GetTagsBySiteID :many
SELECT * FROM site_tags WHERE tag_site = ?;

-- name: GetMetrics :many
SELECT * FROM site_metrics;

-- name: GetMetricsBySiteID :one
SELECT * FROM site_metrics WHERE metric_site = ?;

-- name: GetValidSites :many
SELECT * FROM valid_sites;

-- name: GetValidSitesWithMetrics :many
SELECT * FROM valid_sites_with_metrics;

-- name: GetValidSitesWithMetricsFromMostTotalVisits :many
SELECT *
FROM valid_sites_with_metrics
ORDER BY metric_visits_total DESC, site_id ASC
LIMIT 30;

-- name: GetValidSitesWithMetricsFromMostTotalVisitsLessThan :many
SELECT *
FROM valid_sites_with_metrics
WHERE (metric_visits_total < ?)
OR (metric_visits_total = ? AND site_id > ?)
ORDER BY metric_visits_total DESC, site_id ASC
LIMIT 30;

-- name: GetValidSitesWithMetricsFromLeastTotalVisits :many
SELECT *
FROM valid_sites_with_metrics
ORDER BY metric_visits_total ASC, site_id ASC
LIMIT 30;

-- name: GetValidSitesWithMetricsFromLeastTotalVisitsMoreThan :many
SELECT *
FROM valid_sites_with_metrics
WHERE (metric_visits_total > ?)
OR (metric_visits_total = ? AND site_id > ?)
ORDER BY metric_visits_total ASC, site_id ASC
LIMIT 30;
