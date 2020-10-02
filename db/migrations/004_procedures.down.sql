-- BEGIN;

DROP FUNCTION IF EXISTS sfd.loginUser(userID UUID, loginTime TIMESTAMPTZ);

DROP FUNCTION IF EXISTS sfd.activateUser (userID UUID, approverID UUID);

DROP FUNCTION IF EXISTS sfd.createUser(
    id UUID,
    username VARCHAR(255),
    email VARCHAR(255),
    full_name VARCHAR(255),
    is_admin BOOLEAN,
    active BOOLEAN,
    created_at TIMESTAMPTZ,
    password_hash TEXT);

