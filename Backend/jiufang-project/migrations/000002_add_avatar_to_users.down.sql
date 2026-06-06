-- 000002_add_avatar_to_users.down.sql
-- 回滚脚本：删除users表的avatar字段

ALTER TABLE users 
DROP COLUMN avatar;