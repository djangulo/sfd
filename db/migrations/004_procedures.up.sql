-- BEGIN;

CREATE OR REPLACE FUNCTION sfd.loginUser(userID UUID, loginTime TIMESTAMPTZ) RETURNS SETOF sfd.users_stats AS $$
BEGIN
    UPDATE sfd.users SET (last_login, updated_at) = ($2, $2) WHERE id = $1;
    UPDATE sfd.user_stats SET login_count = login_count + 1 WHERE user_id = $1;
    RETURN QUERY SELECT * FROM sfd.users_stats WHERE id = $1 LIMIT 1;
END;
$$ LANGUAGE plpgsql;



CREATE OR REPLACE FUNCTION sfd.createUser(
    id UUID,
    username VARCHAR(255),
    email VARCHAR(255),
    full_name VARCHAR(255),
    is_admin BOOLEAN,
    active BOOLEAN,
    created_at TIMESTAMPTZ,
    password_hash TEXT) RETURNS SETOF sfd.users_stats AS $$
BEGIN
    INSERT INTO sfd.users (
        id, username, email, full_name, is_admin, active, created_at, password_hash
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

    INSERT INTO sfd.user_stats (
        user_id, login_count, items_created, bids_created, bids_won, created_at
    ) VALUES ($1, 0, 0, 0, 0, $7);

    INSERT INTO sfd.user_preferences (
        rows_per_table, color_theme, user_id
    ) VALUES (10, 1, $1);

    RETURN QUERY SELECT * FROM sfd.users_stats
    WHERE sfd.users_stats.id = $1 LIMIT 1;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION sfd.adminActivateUser (userID UUID, approverID UUID)
    RETURNS VOID AS $$
BEGIN
    UPDATE sfd.users SET active = TRUE WHERE id = $1;
    INSERT INTO sfd.admin_user_approvals (user_id, admin_id) VALUES ($1, $2);
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION sfd.adminApproveItem(itemID UUID, approverID UUID)
    RETURNS VOID AS $$
BEGIN
    UPDATE sfd.items SET admin_approved = TRUE WHERE id = $1;
    INSERT INTO sfd.admin_item_approvals (item_id, admin_id) VALUES ($1, $2);
END;
$$ LANGUAGE plpgsql;

-- CREATE OR REPLACE FUNCTION sfd.placeBid(
--     id UUID,
--     created_at TIMESTAMPTZ,
--     amount NUMERIC(9,2),
--     user_id UUID,
--     item_id UUID) RETURNS SETOF sfd.bids_full AS $$
-- DECLARE
--     item RECORD;
--     winningBid RECORD;
-- BEGIN
    
--     SELECT * INTO item FROM sfd.items WHERE id = $5;
--     -- select winning bid
--     SELECT * INTO winningBid FROM (
--         SELECT * FROM sfd.item_bids WHERE item_id = $5;
--     ) AS tmp_bids LIMIT 1;


--     INSERT INTO sfd.item_bids (
--         id, created_at, amount, user_id, item_id, valid
--     ) VALUES ($1, $2, $3, $4, $5, TRUE);



--     INSERT INTO sfd.users (
--         id, username, email, full_name, is_admin, active, created_at, password_hash
--     ) VALUES ($1, $2, $3, $4, $5, $6, $7, crypt($8, gen_salt('md5')));

--     INSERT INTO sfd.user_stats (
--         user_id, login_count, items_created, bids_created, bids_won, created_at
--     ) VALUES ($1, 0, 0, 0, 0, $7);

--     INSERT INTO sfd.user_preferences (
--         rows_per_table, color_theme, user_id
--     ) VALUES (10, 1, $1);

--     RETURN QUERY SELECT * FROM sfd.users_stats
--     WHERE sfd.users_stats.id = $1 LIMIT 1;
-- END;
-- $$ LANGUAGE plpgsql;

-- COMMIT;

