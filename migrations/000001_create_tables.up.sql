-- Active: 1757133398577@@localhost@5432@utm
-- Create business users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

-- Create accounts table
CREATE TABLE IF NOT EXISTS accounts (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

-- Create users_accounts junction table
CREATE TABLE IF NOT EXISTS users_accounts (
    id BIGSERIAL PRIMARY KEY,
    uid BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    account_id BIGINT NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    UNIQUE (uid, account_id)
);

-- Create auth_users table for system authentication
CREATE TABLE IF NOT EXISTS auth_users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create groups table for RBAC
CREATE TABLE IF NOT EXISTS groups (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT
);

-- Create group_members table
CREATE TABLE IF NOT EXISTS group_members (
    group_id BIGINT NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
    auth_user_id BIGINT NOT NULL REFERENCES auth_users (id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, auth_user_id)
);

-- Create group_permissions table
CREATE TABLE IF NOT EXISTS group_permissions (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
    resource TEXT NOT NULL,
    action TEXT NOT NULL,
    UNIQUE (group_id, resource, action)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_accounts_uid ON users_accounts (uid);

CREATE INDEX IF NOT EXISTS idx_users_accounts_account_id ON users_accounts (account_id);

CREATE INDEX IF NOT EXISTS idx_group_members_auth_user_id ON group_members (auth_user_id);

CREATE INDEX IF NOT EXISTS idx_group_permissions_group_id ON group_permissions (group_id);