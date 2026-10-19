DROP TABLE IF EXISTS oauth_accounts;

-- Note: We do not re-add NOT NULL constraint in rollback to avoid breaking OAuth users with NULL passwords
