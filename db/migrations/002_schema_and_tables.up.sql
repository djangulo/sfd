BEGIN;
CREATE SCHEMA IF NOT EXISTS sfd;
CREATE TABLE IF NOT EXISTS sfd.stampz (
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NULL,
    deleted_at TIMESTAMPTZ NULL
);
CREATE TABLE IF NOT EXISTS sfd.users (
    id UUID PRIMARY KEY,
    username VARCHAR(255) UNIQUE,
    email VARCHAR(255) UNIQUE,
    full_name VARCHAR(255) NULL,
    password_hash TEXT NOT NULL,
    active boolean DEFAULT FALSE,
    profile_public boolean DEFAULT FALSE,
    is_admin boolean DEFAULT FALSE,
    last_login TIMESTAMPTZ NULL
)
INHERITS (
    sfd.stampz
);
CREATE UNIQUE INDEX index_users_on_created_at_id ON sfd.users
    USING btree (created_at, id);

CREATE TABLE IF NOT EXISTS sfd.user_stats (
    user_id UUID PRIMARY KEY REFERENCES sfd.users (id),
    login_count INTEGER,
    bids_created INTEGER,
    bids_won INTEGER,
    items_created INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS sfd.admin_user_approvals (
    admin_id UUID REFERENCES sfd.users (id),
    user_id UUID REFERENCES sfd.users (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT admin_user_approvals_pk PRIMARY KEY (admin_id, user_id)
);



CREATE TABLE IF NOT EXISTS sfd.user_preferences (
    color_theme INTEGER NOT NULL DEFAULT 1,
    language VARCHAR(20) NOT NULL DEFAULT 'en-US',
    user_id UUID REFERENCES sfd.users (id),
    CONSTRAINT user_preferences_pk PRIMARY KEY (user_id)
) INHERITS (sfd.stampz);

CREATE TABLE IF NOT EXISTS sfd.user_phones (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES sfd.users (id),
    number VARCHAR(30)
) INHERITS (sfd.stampz);

CREATE TABLE IF NOT EXISTS sfd.user_addresses (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES sfd.users (id),
    address TEXT,
    kind INTEGER -- billing, shipping
) INHERITS (sfd.stampz);

CREATE TABLE IF NOT EXISTS sfd.files (
    path TEXT NULL,
    abs_path TEXT NULL,
    original_filename TEXT NULL,
    alt_text TEXT NULL,
    file_ext VARCHAR(20) NULL,
    "order" INTEGER NULL
) INHERITS (sfd.stampz);

CREATE TABLE IF NOT EXISTS sfd.profile_pictures (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES sfd.users (id)
) INHERITS (sfd.files);

CREATE TABLE IF NOT EXISTS sfd.items (
    id UUID PRIMARY KEY,
    owner_id UUID REFERENCES sfd.users (id) NOT NULL,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    description TEXT NULL,
    starting_price NUMERIC(9, 2) NULL,
    min_increment NUMERIC(9, 2) NULL,
    max_price NUMERIC(9, 2) NULL,
    published_at TIMESTAMPTZ NULL,
    bid_interval INTEGER NULL, -- interval in seconds
    bid_deadline TIMESTAMPTZ NOT NULL,
    admin_approved BOOLEAN DEFAULT FALSE,
    closed BOOLEAN DEFAULT FALSE,
    blind boolean DEFAULT FALSE,
    user_notified boolean DEFAULT FALSE
)
INHERITS (
    sfd.stampz
);

CREATE UNIQUE INDEX index_items_on_created_at_id ON sfd.items
    USING btree (created_at, id);

CREATE TABLE IF NOT EXISTS sfd.users_item_watch (
    user_id UUID REFERENCES sfd.users (id),
    item_id UUID REFERENCES sfd.items (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT user_item_subscriptions_pk PRIMARY KEY (user_id, item_id)
);

CREATE TABLE IF NOT EXISTS sfd.admin_item_approvals (
    admin_id UUID REFERENCES sfd.users (id),
    item_id UUID REFERENCES sfd.items (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT admin_item_approvals_pk PRIMARY KEY (admin_id, item_id)
);

CREATE TABLE IF NOT EXISTS sfd.item_images (
    id UUID PRIMARY KEY,
    item_id UUID REFERENCES sfd.items (id)
)
INHERITS (
    sfd.files
);
CREATE TABLE IF NOT EXISTS sfd.item_bids (
    id UUID PRIMARY KEY,
    amount NUMERIC(9, 2) NOT NULL,
    user_id UUID NOT NULL REFERENCES sfd.users (id),
    item_id UUID NOT NULL REFERENCES sfd.items (id),
    valid boolean DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NULL
    -- users cannot update or delete their bids, only create new ones
    -- and invalidate already existing ones
);
CREATE UNIQUE INDEX index_item_bids_on_created_at_id ON sfd.item_bids
    USING btree (created_at, id);
CREATE INDEX index_item_bids_on_amount ON sfd.item_bids
    USING btree (amount);

CREATE TABLE IF NOT EXISTS sfd.nomail_list (
    email VARCHAR(255) NOT NULL,
    kind int NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT nomail_list_pk PRIMARY KEY (email, kind)
);
CREATE TABLE IF NOT EXISTS sfd.sessions (
    id VARCHAR(64) PRIMARY KEY,
    values BYTEA NULL,
    expires TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NULL
);
CREATE TABLE IF NOT EXISTS sfd.tokens (
    digest TEXT,
    expires TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ NULL,
    user_id UUID NULL,
    pw_hash TEXT NULL,
    last_login TIMESTAMPTZ NULL,
    kind INTEGER NOT NULL,
    CONSTRAINT token_pk PRIMARY KEY (digest, kind)
);
COMMIT;

