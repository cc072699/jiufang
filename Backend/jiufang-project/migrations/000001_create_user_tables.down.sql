-- 000001_create_user_tables.down.sql
-- 回滚脚本：删除用户组成员关联表和用户表

DROP TABLE IF EXISTS user_group_members;
DROP TABLE IF EXISTS users;