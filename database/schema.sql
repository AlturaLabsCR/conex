-- DDL

CREATE TABLE users (
  user_id INTEGER NOT NULL,
  user_email VARCHAR(63) NOT NULL,
  user_name VARCHAR(63) NOT NULL,
  user_created_unix INTEGER NOT NULL,
  user_modified_unix INTEGER NOT NULL,
  user_deleted INTEGER NOT NULL DEFAULT 0,

  CONSTRAINT pk_users PRIMARY KEY (user_id),
  CONSTRAINT uq_users_email UNIQUE (user_email),
  CONSTRAINT ck_users_deleted CHECK (user_deleted IN (0,1))
);

CREATE TABLE sessions (
  session_id INTEGER NOT NULL,
  session_user INTEGER NOT NULL,
  session_os VARCHAR(7) NOT NULL,
  session_created_unix INTEGER NOT NULL,

  CONSTRAINT pk_sessions PRIMARY KEY (session_id)
);

CREATE TABLE plans (
  plan_id INTEGER NOT NULL,
  plan_name VARCHAR(32) NOT NULL,
  plan_description VARCHAR(1023) NOT NULL,
  plan_amount REAL NOT NULL,
  plan_created_unix INTEGER NOT NULL,
  plan_modified_unix INTEGER NOT NULL,
  plan_active INTEGER NOT NULL DEFAULT 1,

  CONSTRAINT pk_plans PRIMARY KEY (plan_id),
  CONSTRAINT ck_plans_active CHECK (plan_active IN (0,1))
);

CREATE TABLE user_plans (
  user_plan_id INTEGER NOT NULL,
  user_plan_user INTEGER NOT NULL,
  user_plan_plan INTEGER NOT NULL,
  user_plan_created_unix INTEGER NOT NULL,
  user_plan_modified_unix INTEGER NOT NULL,
  user_plan_due_unix INTEGER NOT NULL,
  user_plan_active INTEGER NOT NULL DEFAULT 1,

  CONSTRAINT pk_user_plans PRIMARY KEY (user_plan_id),
  CONSTRAINT fk_user_plans_user FOREIGN KEY (user_plan_user) REFERENCES users(user_id),
  CONSTRAINT fk_user_plans_plan FOREIGN KEY (user_plan_plan) REFERENCES plans(plan_id),
  CONSTRAINT uq_user_plans_active UNIQUE (user_plan_user, user_plan_active),
  CONSTRAINT ck_user_plans_active CHECK (user_plan_active IN (0,1))
);

CREATE TABLE payments (
  payment_id INTEGER NOT NULL,
  payment_user INTEGER NOT NULL,
  payment_plan INTEGER NOT NULL,
  payment_amount REAL NOT NULL,
  payment_date_unix INTEGER NOT NULL,
  payment_successful INTEGER NOT NULL DEFAULT 1,
  payment_reference VARCHAR(255) NOT NULL,

  CONSTRAINT pk_payments PRIMARY KEY (payment_id),
  CONSTRAINT fk_payments_user FOREIGN KEY (payment_user) REFERENCES users(user_id),
  CONSTRAINT fk_payments_plan FOREIGN KEY (payment_plan) REFERENCES plans(plan_id),
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
  site_html_staging TEXT NOT NULL,
  site_created_unix INTEGER NOT NULL,
  site_modified_unix INTEGER NOT NULL,
  site_published INTEGER NOT NULL DEFAULT 1,
  site_deleted INTEGER NOT NULL DEFAULT 0,

  CONSTRAINT pk_sites PRIMARY KEY (site_id),
  CONSTRAINT fk_sites_user FOREIGN KEY (site_user) REFERENCES users(user_id),
  CONSTRAINT uq_sites_slug UNIQUE (site_slug),
  CONSTRAINT ck_sites_deleted CHECK (site_deleted IN (0,1))
);

CREATE TABLE site_metrics (
  metric_id INTEGER NOT NULL,
  metric_site INTEGER NOT NULL,
  metric_visits_total INTEGER NOT NULL,

  CONSTRAINT pk_site_metrics PRIMARY KEY (metric_id),
  CONSTRAINT uq_site_metrics UNIQUE (metric_site)
);

CREATE TABLE site_objects (
  object_id INTEGER NOT NULL,
  object_site INTEGER NOT NULL,
  object_bucket VARCHAR(63) NOT NULL,
  object_key VARCHAR(255) NOT NULL,
  object_mime VARCHAR(63) NOT NULL,
  object_md5 VARCHAR(32) NOT NULL,
  object_size_bytes INTEGER NOT NULL,
  object_created_unix INTEGER NOT NULL,
  object_modified_unix INTEGER NOT NULL,
  object_deleted INTEGER NOT NULL DEFAULT 0,

  CONSTRAINT pk_site_objects PRIMARY KEY (object_id),
  CONSTRAINT fk_site_objects_site FOREIGN KEY (object_site) REFERENCES sites(site_id),
  CONSTRAINT uq_site_objects_address UNIQUE (object_bucket, object_key),
  CONSTRAINT uq_site_objects_md5 UNIQUE (object_bucket, object_md5),
  CONSTRAINT ck_site_objects_deleted CHECK (object_deleted IN (0,1))
);

CREATE TABLE site_banners (
  banner_id INTEGER NOT NULL,
  banner_site INTEGER NOT NULL,
  banner_object INTEGER NOT NULL,

  CONSTRAINT pk_site_banners PRIMARY KEY (banner_id),
  CONSTRAINT uq_site_banners_site UNIQUE (banner_site),
  CONSTRAINT fk_site_banners_site FOREIGN KEY (banner_site) REFERENCES sites(site_id),
  CONSTRAINT fk_site_banners_object FOREIGN KEY (banner_object) REFERENCES site_objects(object_id)
);

CREATE VIEW valid_sites AS
SELECT * FROM sites WHERE
site_published = 1
AND
site_deleted = 0;

CREATE VIEW valid_sites_with_metrics AS
SELECT v.*, m.*
FROM valid_sites AS v INNER JOIN site_metrics AS m
ON v.site_id = m.metric_site;

CREATE INDEX idx_sites_user ON sites(site_user);
CREATE INDEX idx_sites_published_deleted ON sites(site_published, site_deleted);
CREATE INDEX idx_site_metrics_site ON site_metrics(metric_site);
CREATE INDEX idx_site_metrics_visits_total ON site_metrics(metric_visits_total);
-- CREATE INDEX idx_sites_created_unix ON sites(site_created_unix DESC); -- Sites by creation date
-- CREATE INDEX idx_sites_modified_unix ON sites(site_modified_unix DESC); -- Last updated sites
