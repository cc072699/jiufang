-- 000003_create_permission_tables.down.sql
-- 回滚脚本：删除用户组和权限表

-- 删除权限表
DROP TABLE IF EXISTS permissions;

-- 删除用户组表
DROP TABLE IF EXISTS user_groups;