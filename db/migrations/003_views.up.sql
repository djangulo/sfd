BEGIN;
CREATE VIEW sfd.users_stats AS
SELECT
    users.id,
    users.username,
    users.email,
    users.full_name,
    users.active,
    users.profile_public,
    users.is_admin,
    users.last_login,
    stats.login_count AS "stats.login_count",
    stats.items_created AS "stats.items_created",
    stats.bids_created AS "stats.bids_created",
    stats.bids_won AS "stats.bids_won",
    pictures.id AS "picture.id",
    pictures.path AS "picture.path",
    pictures.abs_path AS "picture.abs_path",
    pictures.original_filename AS "picture.original_filename",
    pictures.file_ext AS "picture.file_ext",
    pictures.alt_text AS "picture.alt_text",
    pictures.created_at AS "picture.created_at",
    pictures.updated_at AS "picture.updated_at",
    preferences.language AS "preferences.language",
    preferences.color_theme AS "preferences.color_theme"
FROM
    sfd.users AS users
    LEFT JOIN sfd.user_stats AS stats ON users.id = stats.user_id
    LEFT JOIN sfd.profile_pictures AS pictures ON users.id = pictures.user_id
    LEFT JOIN sfd.user_preferences AS preferences ON users.id = preferences.user_id;
CREATE VIEW sfd.items_owner AS
SELECT
    items.*,
    users.id AS "owner.id",
    users.username AS "owner.username",
    users.email AS "owner.email",
    users.full_name AS "owner.full_name",
    users.created_at AS "owner.created_at",
    users.updated_at AS "owner.updated_at",
    users.deleted_at AS
    "owner.deleted_at",
    stats.login_count AS "owner.stats.login_count",
    stats.items_created AS "owner.stats.items_created",
    stats.bids_created AS "owner.stats.bids_created",
    item_images.id AS "cover_image.id",
    item_images.path AS "cover_image.path",
    item_images.abs_path AS "cover_image.abs_path",
    item_images.original_filename AS "cover_image.original_filename",
    item_images.alt_text AS "cover_image.alt_text",
    item_images.file_ext AS "cover_image.file_ext",
    item_images.order AS "cover_image.order"
FROM
    sfd.items AS items
    LEFT JOIN sfd.users AS users ON items.owner_id = users.id
    LEFT JOIN sfd.user_stats AS stats ON users.id = stats.user_id
    LEFT JOIN (
        SELECT
            id,
            path,
            abs_path,
            original_filename,
            alt_text,
            file_ext,
            created_at,
            updated_at,
            "order",
            item_id
        FROM
            sfd.item_images
        WHERE
            "order" = 1) AS item_images ON item_images.item_id = items.id;
CREATE VIEW sfd.bids_full AS
SELECT
    bids.*,
    items.id AS "item.id",
    items.owner_id AS "item.owner_id",
    items.name AS "item.name",
    items.slug AS "item.slug",
    items.description AS "item.description",
    items.starting_price AS "item.starting_price",
    items.max_price AS "item.max_price",
    items.min_increment AS "item.min_increment",
    items.bid_interval AS "item.bid_interval",
    items.bid_deadline AS "item.bid_deadline",
    items.blind AS "item.blind",
    items.closed AS "item.closed",
    items.created_at AS "item.created_at",
    items.updated_at AS "item.updated_at",
    items.deleted_at AS
    "item.deleted_at",
    items.published_at AS "item.published_at",
    items.admin_approved AS "item.admin_approved",
    "owner.username" AS "item.owner.username",
    "owner.email" AS "item.owner.email",
    "owner.full_name" AS "item.owner.full_name",
    "owner.created_at" AS "item.owner.created_at",
    "owner.updated_at" AS "item.owner.updated_at",
    "owner.deleted_at" AS
    "item.owner.deleted_at",
    "owner.stats.login_count" AS "item.owner.stats.login_count",
    "owner.stats.items_created" AS "item.owner.stats.items_created",
    "owner.stats.bids_created" AS "item.owner.stats.bids_created",
    "cover_image.id" AS "item.cover_image.id",
    "cover_image.path" AS "item.cover_image.path",
    "cover_image.abs_path" AS "item.cover_image.abs_path",
    "cover_image.original_filename" AS "item.cover_image.original_filename",
    "cover_image.alt_text" AS "item.cover_image.alt_text",
    "cover_image.file_ext" AS "item.cover_image.file_ext",
    "cover_image.order" AS "item.cover_image.order",
    users.id AS "user.id",
    users.username AS "user.username",
    users.email AS "user.email",
    users.full_name AS "user.full_name",
    users.created_at AS "user.created_at",
    users.updated_at AS "user.updated_at",
    users.deleted_at AS
    "user.deleted_at",
    stats.login_count AS "user.stats.login_count",
    stats.items_created AS "user.stats.items_created",
    stats.bids_created AS "user.stats.bids_created"
FROM
    sfd.item_bids AS bids
    LEFT JOIN sfd.items_owner AS items ON bids.item_id = items.id
    LEFT JOIN sfd.users AS users ON bids.user_id = users.id
    LEFT JOIN sfd.user_stats AS stats ON users.id = stats.user_id
    ORDER BY amount DESC;
-- CREATE VIEW user_profile_view_full AS
-- SELECT
--     *
-- FROM
--     sfd.accounts AS USER
--     LEFT JOIN sfd.user_phones AS phones ON user.id = phones.user_id
--     LEFT JOIN sfd.profile_pictures AS ppic ON user.id = ppic.user_id;
-- CREATE VIEW user_profile_view AS
-- SELECT
--     user.id AS id,
--     user.username AS username,
--     user.email AS email,
--     user.full_name AS full_name,
--     user.active AS active,
--     user.is_admin AS is_admin,
--     user.approving_admin_id AS approving_admin_id,
--     user.phone_number AS phone_number ppic.id AS picture_id,
--     ppic.abs_path AS picture_abs_path,
--     ppic.original_filename AS picture_original_filename,
--     ppic.alt_text AS picture_alt_text
-- FROM
--     user_profile_view_full;

COMMIT;

