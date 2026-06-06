-- 000002_add_avatar_to_users.up.sql
-- 为users表新增avatar字段（PostgreSQL版本）

ALTER TABLE users 
ADD COLUMN avatar VARCHAR(255) DEFAULT NULL;

-- 添加字段注释
COMMENT ON COLUMN users.avatar IS '用户头像URL';