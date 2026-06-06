-- 000001_create_user_tables.up.sql
-- 创建用户表（PostgreSQL版本）
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    snowflake_id BIGINT NOT NULL UNIQUE,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    role VARCHAR(20) NOT NULL,
    status INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL
);

-- 创建索引
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- 创建用户组成员关联表
CREATE TABLE user_group_members (
    id BIGSERIAL PRIMARY KEY,
    snowflake_id BIGINT NOT NULL UNIQUE,
    user_id BIGINT NOT NULL,
    group_id BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_user_group_members_user_id ON user_group_members(user_id);
CREATE INDEX idx_user_group_members_group_id ON user_group_members(group_id);
CREATE UNIQUE INDEX idx_user_group_members_unique ON user_group_members(user_id, group_id);

-- 添加表注释（PostgreSQL使用COMMENT命令）
COMMENT ON TABLE users IS '用户表';
COMMENT ON COLUMN users.id IS '物理主键(自增ID)';
COMMENT ON COLUMN users.snowflake_id IS '业务唯一标识(雪花ID)';
COMMENT ON COLUMN users.username IS '用户名';
COMMENT ON COLUMN users.password IS '密码(bcrypt加密)';
COMMENT ON COLUMN users.email IS '邮箱';
COMMENT ON COLUMN users.role IS '角色(admin/manager/executive)';
COMMENT ON COLUMN users.status IS '用户状态(1启用/0停用)';
COMMENT ON COLUMN users.created_at IS '创建时间';
COMMENT ON COLUMN users.updated_at IS '更新时间';
COMMENT ON COLUMN users.deleted_at IS '删除时间(软删除)';

COMMENT ON TABLE user_group_members IS '用户组成员关联表';
COMMENT ON COLUMN user_group_members.id IS '物理主键(自增ID)';
COMMENT ON COLUMN user_group_members.snowflake_id IS '业务唯一标识(雪花ID)';
COMMENT ON COLUMN user_group_members.user_id IS '用户ID(外键)';
COMMENT ON COLUMN user_group_members.group_id IS '用户组ID(外键)';
COMMENT ON COLUMN user_group_members.created_at IS '创建时间';