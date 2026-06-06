ALTER TABLE users ADD COLUMN is_first_login BOOLEAN NOT NULL DEFAULT TRUE;

COMMENT ON COLUMN users.is_first_login IS '是否首次登录（初始密码状态）';