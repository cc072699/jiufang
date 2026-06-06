-- 000003_create_permission_tables.up.sql
-- 创建用户组和权限表（PostgreSQL版本）

-- 创建用户组表
CREATE TABLE user_groups (
    id BIGSERIAL PRIMARY KEY,
    snowflake_id BIGINT NOT NULL UNIQUE,
    name VARCHAR(50) NOT NULL UNIQUE,
    description VARCHAR(200) DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL
);

-- 创建索引
CREATE INDEX idx_user_groups_deleted_at ON user_groups(deleted_at);

-- 创建权限表
CREATE TABLE permissions (
    id BIGSERIAL PRIMARY KEY,
    snowflake_id BIGINT NOT NULL UNIQUE,
    group_id BIGINT NOT NULL,
    resource_type VARCHAR(20) NOT NULL,
    resource_name VARCHAR(100) NOT NULL,
    permission_action VARCHAR(20) NOT NULL,
    filter_condition VARCHAR(500) DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_permissions_group_id ON permissions(group_id);
CREATE INDEX idx_permissions_resource ON permissions(resource_type, resource_name);

-- 添加表注释
COMMENT ON TABLE user_groups IS '用户组表';
COMMENT ON COLUMN user_groups.id IS '物理主键(自增ID)';
COMMENT ON COLUMN user_groups.snowflake_id IS '业务唯一标识(雪花ID)';
COMMENT ON COLUMN user_groups.name IS '用户组名称';
COMMENT ON COLUMN user_groups.description IS '用户组描述';
COMMENT ON COLUMN user_groups.created_at IS '创建时间';
COMMENT ON COLUMN user_groups.updated_at IS '更新时间';
COMMENT ON COLUMN user_groups.deleted_at IS '删除时间(软删除)';

COMMENT ON TABLE permissions IS '权限表';
COMMENT ON COLUMN permissions.id IS '物理主键(自增ID)';
COMMENT ON COLUMN permissions.snowflake_id IS '业务唯一标识(雪花ID)';
COMMENT ON COLUMN permissions.group_id IS '用户组ID(外键)';
COMMENT ON COLUMN permissions.resource_type IS '资源类型(table/field/operation)';
COMMENT ON COLUMN permissions.resource_name IS '资源名称(表名/字段名/操作名)';
COMMENT ON COLUMN permissions.permission_action IS '权限动作(read/write/export)';
COMMENT ON COLUMN permissions.filter_condition IS '数据级过滤条件(SQL WHERE片段)';
COMMENT ON COLUMN permissions.created_at IS '创建时间';
COMMENT ON COLUMN permissions.updated_at IS '更新时间';

-- 插入预置用户组（PRD BR-032要求）
INSERT INTO user_groups (snowflake_id, name, description, created_at, updated_at) VALUES
(1000000000000000001, '管理员组', '系统管理员，拥有全部权限', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(1000000000000000002, '高管组', '公司高管，可查询全公司数据', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(1000000000000000003, '采购组', '采购部门，可查询采购相关数据', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(1000000000000000004, '销售组', '销售部门，可查询销售相关数据', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(1000000000000000005, '财务组', '财务部门，可查询财务相关数据', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);