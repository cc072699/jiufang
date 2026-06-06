-- 删除触发器
DROP TRIGGER IF EXISTS update_scheduled_reports_updated_at ON scheduled_reports;

-- 删除触发器函数
DROP FUNCTION IF EXISTS update_updated_at_column();

-- 删除推送记录表
DROP TABLE IF EXISTS push_records;

-- 删除定时报告表
DROP TABLE IF EXISTS scheduled_reports;