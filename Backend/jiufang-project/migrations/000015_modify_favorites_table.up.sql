-- 修改收藏表结构
-- 1. query_record_id 改为可选
-- 2. 删除 title 字段
-- 3. 增加 name、input、sql、description 字段

-- 删除旧的唯一索引
DROP INDEX IF EXISTS idx_favorites_unique;

-- 修改 query_record_id 为可选
ALTER TABLE favorites ALTER COLUMN query_record_id DROP NOT NULL;

-- 删除 title 字段
ALTER TABLE favorites DROP COLUMN IF EXISTS title;

-- 增加 name 字段
ALTER TABLE favorites ADD COLUMN name VARCHAR(200) NOT NULL DEFAULT '';

-- 增加 input 字段
ALTER TABLE favorites ADD COLUMN input TEXT NOT NULL DEFAULT '';

-- 增加 sql 字段
ALTER TABLE favorites ADD COLUMN sql TEXT NOT NULL DEFAULT '';

-- 增加 description 字段
ALTER TABLE favorites ADD COLUMN description TEXT DEFAULT NULL;

-- 为 name 字段创建索引
CREATE INDEX idx_favorites_name ON favorites(name);

-- 更新注释
COMMENT ON COLUMN favorites.query_record_id IS '查询记录ID(可选，用于关联历史记录)';
COMMENT ON COLUMN favorites.name IS '收藏名称';
COMMENT ON COLUMN favorites.input IS '用户输入的自然语言问题';
COMMENT ON COLUMN favorites.sql IS 'SQL语句';
COMMENT ON COLUMN favorites.description IS '收藏描述';