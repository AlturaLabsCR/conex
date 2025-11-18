-- DDL

CREATE TABLE users (
  user_id BIGSERIAL PRIMARY KEY,
  user_email VARCHAR(63) NOT NULL,
  user_created_unix BIGINT NOT NULL,
  user_modified_unix BIGINT NOT NULL,
  user_deleted BIGINT NOT NULL DEFAULT 0,
  CONSTRAINT ck_users_deleted CHECK (user_deleted IN (0,1))
);

CREATE UNIQUE INDEX uq_users_email_active_only
ON users(user_email)
WHERE user_email <> '' AND user_deleted = 0;

CREATE TABLE sessions (
  session_id BIGSERIAL PRIMARY KEY,
  "session_user" BIGINT NOT NULL,
  session_device VARCHAR(63) NOT NULL,
  session_last_login_unix BIGINT NOT NULL
);

CREATE TABLE user_plans (
  user_plan_id BIGSERIAL PRIMARY KEY,
  user_plan_user BIGINT NOT NULL,
  user_plan_created_unix BIGINT NOT NULL,
  user_plan_modified_unix BIGINT NOT NULL,
  user_plan_due_unix BIGINT NOT NULL,
  user_plan_active BIGINT NOT NULL DEFAULT 1,
  CONSTRAINT fk_user_plans_user FOREIGN KEY (user_plan_user) REFERENCES users(user_id),
  CONSTRAINT uq_user_plans_user UNIQUE (user_plan_user),
  CONSTRAINT ck_user_plans_active CHECK (user_plan_active IN (0,1))
);

CREATE TABLE payments (
  payment_id BIGSERIAL PRIMARY KEY,
  payment_user BIGINT NOT NULL,
  payment_amount REAL NOT NULL,
  payment_date_unix BIGINT NOT NULL,
  payment_successful BIGINT NOT NULL DEFAULT 1,
  payment_reference VARCHAR(255) NOT NULL,
  CONSTRAINT fk_payments_user FOREIGN KEY (payment_user) REFERENCES users(user_id),
  CONSTRAINT ck_payments_successful CHECK (payment_successful IN (0,1))
);

CREATE TABLE sites (
  site_id BIGSERIAL PRIMARY KEY,
  site_user BIGINT NOT NULL,
  site_slug VARCHAR(63) NOT NULL,
  site_title VARCHAR(63) NOT NULL,
  site_tags_json VARCHAR(255) NOT NULL,
  site_description VARCHAR(255) NOT NULL,
  site_html_gz BYTEA NOT NULL,
  site_created_unix BIGINT NOT NULL,
  site_modified_unix BIGINT NOT NULL,
  site_home_page BIGINT NOT NULL DEFAULT 0,
  site_published BIGINT NOT NULL DEFAULT 1,
  site_deleted BIGINT NOT NULL DEFAULT 0,
  CONSTRAINT fk_sites_user FOREIGN KEY (site_user) REFERENCES users(user_id),
  CONSTRAINT uq_sites_slug UNIQUE (site_slug),
  CONSTRAINT ck_sites_deleted CHECK (site_deleted IN (0,1))
);

CREATE TABLE site_sync (
  site_sync_id BIGINT PRIMARY KEY,
  site_sync_data_gz BYTEA NOT NULL,
  site_sync_last_update_unix BIGINT NOT NULL,
  CONSTRAINT fk_site_sync FOREIGN KEY (site_sync_id) REFERENCES sites(site_id) ON DELETE CASCADE
);


CREATE TABLE site_metrics (
  metric_id BIGSERIAL PRIMARY KEY,
  metric_site BIGINT NOT NULL,
  metric_visits_total BIGINT NOT NULL,
  CONSTRAINT fk_sites_metrics_site FOREIGN KEY (metric_site) REFERENCES sites(site_id) ON DELETE CASCADE,
  CONSTRAINT uq_site_metrics UNIQUE (metric_site)
);

CREATE TABLE site_objects (
  object_id BIGSERIAL PRIMARY KEY,
  object_site BIGINT NOT NULL DEFAULT 0,
  object_bucket VARCHAR(63) NOT NULL,
  object_key VARCHAR(255) NOT NULL,
  object_mime VARCHAR(63) NOT NULL,
  object_md5 VARCHAR(32) NOT NULL,
  object_size_bytes BIGINT NOT NULL,
  object_created_unix BIGINT NOT NULL,
  object_modified_unix BIGINT NOT NULL,
  CONSTRAINT fk_site_objects_site FOREIGN KEY (object_site) REFERENCES sites(site_id) ON DELETE CASCADE,
  CONSTRAINT uq_site_objects_address UNIQUE (object_bucket, object_key),
  CONSTRAINT uq_site_objects_md5 UNIQUE (object_bucket, object_md5)
);

CREATE TABLE site_banners (
  banner_id BIGSERIAL PRIMARY KEY,
  banner_site BIGINT NOT NULL,
  banner_object BIGINT NOT NULL,
  CONSTRAINT uq_site_banners_site UNIQUE (banner_site),
  CONSTRAINT fk_site_banners_site FOREIGN KEY (banner_site) REFERENCES sites(site_id) ON DELETE CASCADE,
  CONSTRAINT fk_site_banners_object FOREIGN KEY (banner_object) REFERENCES site_objects(object_id)
);

CREATE VIEW sites_with_metrics AS
SELECT s.*, m.*
FROM sites AS s INNER JOIN site_metrics AS m
ON s.site_id = m.metric_site;

CREATE INDEX idx_sites_user ON sites(site_user);
CREATE INDEX idx_sites_published_deleted ON sites(site_published, site_deleted);
CREATE INDEX idx_site_metrics_site ON site_metrics(metric_site);
CREATE INDEX idx_site_metrics_visits_total ON site_metrics(metric_visits_total);
