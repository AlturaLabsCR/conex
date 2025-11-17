-- DDL

CREATE TABLE users (
  user_id INTEGER NOT NULL,
  user_email VARCHAR(63) NOT NULL,
  user_created_unix INTEGER NOT NULL,
  user_modified_unix INTEGER NOT NULL,
  user_deleted INTEGER NOT NULL DEFAULT 0,

  CONSTRAINT pk_users PRIMARY KEY (user_id),
  CONSTRAINT ck_users_deleted CHECK (user_deleted IN (0,1))
);

CREATE UNIQUE INDEX uq_users_email_active_only
ON users(user_email)
WHERE user_email <> '' AND user_deleted = 0;

CREATE TABLE sessions (
  session_id INTEGER NOT NULL,
  session_user INTEGER NOT NULL,
  session_device VARCHAR(7) NOT NULL,
  session_last_login_unix INTEGER NOT NULL,

  CONSTRAINT pk_sessions PRIMARY KEY (session_id)
);

CREATE TABLE user_plans (
  user_plan_id INTEGER NOT NULL,
  user_plan_user INTEGER NOT NULL,
  user_plan_created_unix INTEGER NOT NULL,
  user_plan_modified_unix INTEGER NOT NULL,
  user_plan_due_unix INTEGER NOT NULL,
  user_plan_active INTEGER NOT NULL DEFAULT 1,

  CONSTRAINT pk_user_plans PRIMARY KEY (user_plan_id),
  CONSTRAINT fk_user_plans_user FOREIGN KEY (user_plan_user) REFERENCES users(user_id),
  CONSTRAINT uq_user_plans_user UNIQUE (user_plan_user),
  CONSTRAINT ck_user_plans_active CHECK (user_plan_active IN (0,1))
);

CREATE TABLE payments (
  payment_id INTEGER NOT NULL,
  payment_user INTEGER NOT NULL,
  payment_amount REAL NOT NULL,
  payment_date_unix INTEGER NOT NULL,
  payment_successful INTEGER NOT NULL DEFAULT 1,
  payment_reference VARCHAR(255) NOT NULL,

  CONSTRAINT pk_payments PRIMARY KEY (payment_id),
  CONSTRAINT fk_payments_user FOREIGN KEY (payment_user) REFERENCES users(user_id),
  CONSTRAINT ck_payments_successful CHECK (payment_successful IN (0,1))
);

CREATE TABLE sites (
  site_id INTEGER NOT NULL,
  site_user INTEGER NOT NULL,
  site_slug VARCHAR(63) NOT NULL,
  site_title VARCHAR(63) NOT NULL,
  site_tags_json VARCHAR(255) NOT NULL,
  site_description VARCHAR(255) NOT NULL,
  site_html_published TEXT NOT NULL,
  site_created_unix INTEGER NOT NULL,
  site_modified_unix INTEGER NOT NULL,
  site_home_page INTEGER NOT NULL DEFAULT 0,
  site_published INTEGER NOT NULL DEFAULT 1,
  site_deleted INTEGER NOT NULL DEFAULT 0,

  CONSTRAINT pk_sites PRIMARY KEY (site_id),
  CONSTRAINT fk_sites_user FOREIGN KEY (site_user) REFERENCES users(user_id),
  CONSTRAINT uq_sites_slug UNIQUE (site_slug),
  CONSTRAINT ck_sites_deleted CHECK (site_deleted IN (0,1))
);

CREATE TABLE site_sync (
  site_sync_id INTEGER NOT NULL,
  site_sync_data_staging TEXT NOT NULL,
  site_sync_last_update_unix INTEGER NOT NULL,

  CONSTRAINT pk_site_sync PRIMARY KEY (site_sync_id),
  CONSTRAINT fk_site_sync FOREIGN KEY (site_sync_id) REFERENCES sites(site_id) ON DELETE CASCADE
);

CREATE TABLE site_metrics (
  metric_id INTEGER NOT NULL,
  metric_site INTEGER NOT NULL,
  metric_visits_total INTEGER NOT NULL,

  CONSTRAINT pk_site_metrics PRIMARY KEY (metric_id),
  CONSTRAINT fk_sites_metrics_site FOREIGN KEY (metric_site) REFERENCES sites(site_id) ON DELETE CASCADE,
  CONSTRAINT uq_site_metrics UNIQUE (metric_site)
);

CREATE TABLE site_objects (
  object_id INTEGER NOT NULL,
  object_site INTEGER NOT NULL DEFAULT 0,
  object_bucket VARCHAR(63) NOT NULL,
  object_key VARCHAR(255) NOT NULL,
  object_mime VARCHAR(63) NOT NULL,
  object_md5 VARCHAR(32) NOT NULL,
  object_size_bytes INTEGER NOT NULL,
  object_created_unix INTEGER NOT NULL,
  object_modified_unix INTEGER NOT NULL,

  CONSTRAINT pk_site_objects PRIMARY KEY (object_id),
  CONSTRAINT fk_site_objects_site FOREIGN KEY (object_site) REFERENCES sites(site_id) ON DELETE CASCADE,
  CONSTRAINT uq_site_objects_address UNIQUE (object_bucket, object_key),
  CONSTRAINT uq_site_objects_md5 UNIQUE (object_bucket, object_md5)
);

CREATE TABLE site_banners (
  banner_id INTEGER NOT NULL,
  banner_site INTEGER NOT NULL,
  banner_object INTEGER NOT NULL,

  CONSTRAINT pk_site_banners PRIMARY KEY (banner_id),
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
-- CREATE INDEX idx_sites_created_unix ON sites(site_created_unix DESC); -- Sites by creation date
-- CREATE INDEX idx_sites_modified_unix ON sites(site_modified_unix DESC); -- Last updated sites
