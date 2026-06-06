-- Rollback: Restore scheduled_reports table to original structure

-- Step 1: Add back old columns
ALTER TABLE scheduled_reports
ADD COLUMN query_condition TEXT,
ADD COLUMN schedule_cron VARCHAR(100),
ADD COLUMN push_targets TEXT;

-- Step 2: Migrate data back
-- Convert sql to query_condition
UPDATE scheduled_reports
SET query_condition = sql;

-- Convert schedule_type and schedule_time back to schedule_cron
-- Simplified conversion: daily at 08:00 -> "0 0 8 * * *"
UPDATE scheduled_reports
SET schedule_cron = CASE
    WHEN schedule_type = 'daily' THEN
        CONCAT('0 0 ', SUBSTRING(schedule_time FROM 1 FOR 2), ' * * *')
    WHEN schedule_type = 'weekly' THEN
        CONCAT('0 0 ', SUBSTRING(schedule_time FROM 1 FOR 2), ' * * 0')
    WHEN schedule_type = 'monthly' THEN
        CONCAT('0 0 ', SUBSTRING(schedule_time FROM 1 FOR 2), ' 1 * *')
    ELSE '0 0 8 * * *'
END;

-- Convert recipients to push_targets
UPDATE scheduled_reports
SET push_targets = recipients;

-- Step 3: Drop new columns
ALTER TABLE scheduled_reports
DROP COLUMN sql,
DROP COLUMN schedule_type,
DROP COLUMN schedule_time,
DROP COLUMN recipients;

-- Step 4: Restore constraints
ALTER TABLE scheduled_reports
ALTER COLUMN query_condition SET NOT NULL,
ALTER COLUMN schedule_cron SET NOT NULL;

-- Remove comments
COMMENT ON COLUMN scheduled_reports.sql IS NULL;
COMMENT ON COLUMN scheduled_reports.schedule_type IS NULL;
COMMENT ON COLUMN scheduled_reports.schedule_time IS NULL;
COMMENT ON COLUMN scheduled_reports.recipients IS NULL;