INSERT INTO users (user_id, user_email, user_name, user_created_unix, user_modified_unix, user_deleted)
VALUES
(1, 'alice@example.com', 'Alice', 1729459200, 1729459200, 0),
(2, 'bob@example.com',   'Bob',   1729459200, 1729459200, 0),
(3, 'carol@example.com', 'Carol', 1729459200, 1729459200, 0);

INSERT INTO plans (plan_id, plan_name, plan_description, plan_amount, plan_created_unix, plan_modified_unix, plan_active)
VALUES
(1, 'Free',    'Basic plan with limited features.', 0.00, 1729459200, 1729459200, 1),
(2, 'Pro',     'Professional plan for creators.',   12.99, 1729459200, 1729459200, 1),
(3, 'Business','Advanced plan with analytics.',     29.99, 1729459200, 1729459200, 1);

INSERT INTO user_plans (user_plan_id, user_plan_user, user_plan_plan, user_plan_created_unix, user_plan_modified_unix, user_plan_due_unix, user_plan_active)
VALUES
(1, 1, 2, 1729459200, 1729459200, 1732051200, 1), -- Alice: Pro plan
(2, 2, 1, 1729459200, 1729459200, 1732051200, 1), -- Bob: Free plan
(3, 3, 3, 1729459200, 1729459200, 1732051200, 1); -- Carol: Business plan

INSERT INTO sites (
  site_id, site_user, site_slug, site_title, site_description,
  site_html_published, site_html_staging,
  site_created_unix, site_modified_unix, site_published, site_deleted
)
VALUES
(1, 1, "alice-blog", "Alice's Blog", "Personal thoughts and design updates.",
  "<h1>Welcome to Alice's Blog</h1>", "<h1>Draft: Alice's Blog</h1>",
  1729459200, 1729459200, 1, 0),

(2, 1, "alice-portfolio", "Alice's Portfolio", "Showcasing creative work.",
  "<h1>Alice's Portfolio</h1>", "<h1>Portfolio Draft</h1>",
  1729459200, 1729459200, 1, 0),

(3, 2, "bob-coding", "Bob's Coding Corner", "Programming tutorials and guides.",
  "<h1>Learn to Code with Bob</h1>", "<h1>Draft: Coding Corner</h1>",
  1729459200, 1729459200, 1, 0),

(4, 3, "carol-photography", "Carol's Photography", "Travel and lifestyle photography portfolio.",
  "<h1>Photography by Carol</h1>", "<h1>Draft Photography Page</h1>",
  1729459200, 1729459200, 1, 0);

INSERT INTO site_tags (tag_id, tag_site, tag_name, tag_color_hex)
VALUES
(1, 1, 'design', '#ff3366'),
(2, 1, 'personal', '#33cc99'),
(3, 2, 'portfolio', '#3366ff'),
(4, 3, 'coding', '#ffcc00'),
(5, 3, 'tutorials', '#cc33ff'),
(6, 4, 'photography', '#ff6699'),
(7, 4, 'travel', '#33ccff');

INSERT INTO site_metrics (metric_id, metric_site, metric_visits_total)
VALUES
(1, 1, 1250),
(2, 2, 980),
(3, 3, 4120),
(4, 4, 2210);
