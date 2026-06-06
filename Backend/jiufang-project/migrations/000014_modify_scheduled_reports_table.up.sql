-- Modify scheduled_reports table to align with detailed design document
-- Changes:
-- 1. Rename query_condition to sql
-- 2. Replace schedule_cron with schedule_type and schedule_time
-- 3. Rename push_targets to recipients

-- Step 1: Add new columns
ALTER TABLE scheduled_reports
ADD COLUMN sql TEXT,
ADD COLUMN schedule_type VARCHAR(20),
ADD COLUMN schedule_time VARCHAR(50),
ADD COLUMN recipients TEXT;

-- Step 2: Migrate data
-- Convert query_condition to sql
UPDATE scheduled_reports
SET sql = query_condition;

-- Convert schedule_cron to schedule_type and schedule_time
-- Example: "0 0 8 * * *" (daily at 8:00) -> schedule_type='daily', schedule_time='08:00:00'
-- This is a simplified migration, complex cron expressions may need manual adjustment
UPDATE scheduled_reports
SET
    schedule_type = CASE
        WHEN schedule_cron LIKE '0 0 * * *' THEN 'daily'
        WHEN schedule_cron LIKE '0 0 * * 0' THEN 'weekly'
        WHEN schedule_cron LIKE '0 0 1 * *' THEN 'monthly'
        ELSE 'daily'
    END,
    schedule_time = CASE
        WHEN schedule_cron LIKE '0 0 % * * *' THEN
            CONCAT('0', SUBSTRING(schedule_cron FROM 8 FOR 1), ':00:00')
        ELSE '08:00:00'
    END;

-- Convert push_targets to recipients (JSON array to JSON array)
UPDATE scheduled_reports
SET recipients = push_targets;

-- Step 3: Drop old columns
ALTER TABLE scheduled_reports
DROP COLUMN query_condition,
DROP COLUMN schedule_cron,
DROP COLUMN push_targets;

-- Step 4: Add constraints
ALTER TABLE scheduled_reports
ALTER COLUMN sql SET NOT NULL,
ALTER COLUMN schedule_type SET NOT NULL,
ALTER COLUMN schedule_time SET NOT NULL;

-- Add comments
COMMENT ON COLUMN scheduled_reports.sql IS 'SQL语句';
COMMENT ON COLUMN scheduled_reports.schedule_type IS '定时类型（daily/weekly/monthly）';
COMMENT ON COLUMN scheduled_reports.schedule_time IS '定时时间（ISO8601格式或时间表达式）';
COMMENT ON COLUMN scheduled_reports.recipients IS '接收者列表（JSON数组格式）';