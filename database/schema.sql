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

CREATE TABLE sites (
  site_id INTEGER NOT NULL,
  site_user INTEGER NOT NULL,
  site_slug VARCHAR(63) NOT NULL,
  site_title VARCHAR(63) NOT NULL,
  site_description VARCHAR(255) NOT NULL,
  site_tags_json VARCHAR(255) NOT NULL,
  site_data TEXT NOT NULL,
  site_created_unix INTEGER NOT NULL,
  site_modified_unix INTEGER NOT NULL,
  site_published INTEGER NOT NULL DEFAULT 1,
  site_deleted INTEGER NOT NULL DEFAULT 0,

  CONSTRAINT pk_sites PRIMARY KEY (site_id),
  CONSTRAINT fk_sites_user FOREIGN KEY (site_user) REFERENCES users(user_id),
  CONSTRAINT uq_sites_slug UNIQUE (site_slug),
  CONSTRAINT ck_sites_deleted CHECK (site_deleted IN (0,1))
);

CREATE TABLE site_objects (
  object_id INTEGER NOT NULL,
  object_site INTEGER NOT NULL,
  object_bucket VARCHAR(63) NOT NULL,
  object_key VARCHAR(255) NOT NULL,
  object_mime VARCHAR(63) NOT NULL,
  object_size_bytes INTEGER NOT NULL,
  object_created_unix INTEGER NOT NULL,
  object_modified_unix INTEGER NOT NULL,
  object_deleted INTEGER NOT NULL DEFAULT 0,

  CONSTRAINT pk_site_objects PRIMARY KEY (object_id),
  CONSTRAINT fk_site_objects_site FOREIGN KEY (object_site) REFERENCES sites(site_id),
  CONSTRAINT uq_site_objects_address UNIQUE (object_bucket, object_key),
  CONSTRAINT ck_site_objects_deleted CHECK (object_deleted IN (0,1))
);
