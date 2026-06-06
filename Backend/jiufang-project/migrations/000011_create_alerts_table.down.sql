-- 删除触发器
DROP TRIGGER IF EXISTS update_alerts_updated_at ON alerts;

-- 删除预警规则表
DROP TABLE IF EXISTS alerts;