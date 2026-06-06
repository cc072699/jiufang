-- 回滚收藏表结构修改

-- 删除新增的字段
ALTER TABLE favorites DROP COLUMN IF EXISTS description;
ALTER TABLE favorites DROP COLUMN IF EXISTS sql;
ALTER TABLE favorites DROP COLUMN IF EXISTS input;
ALTER TABLE favorites DROP COLUMN IF EXISTS name;

-- 恢复 title 字段
ALTER TABLE favorites ADD COLUMN title VARCHAR(100) DEFAULT NULL;

-- 恢复 query_record_id 为必填
ALTER TABLE favorites ALTER COLUMN query_record_id SET NOT NULL;

-- 删除 name 字段的索引
DROP INDEX IF EXISTS idx_favorites_name;

-- 恢复旧的唯一索引
CREATE UNIQUE INDEX idx_favorites_unique ON favorites(user_id, query_record_id);

-- 恢复注释
COMMENT ON COLUMN favorites.query_record_id IS '查询记录ID';
COMMENT ON COLUMN favorites.title IS '收藏标题(用户自定义)';