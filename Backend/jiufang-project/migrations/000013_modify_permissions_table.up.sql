-- Modify permissions table to align with detailed design document
-- Changes:
-- 1. Rename resource_name to table_name
-- 2. Add allowed_fields column (JSON array format)
-- 3. Remove resource_type and permission_action columns (not defined in detailed design)

-- Step 1: Add new columns
ALTER TABLE permissions
ADD COLUMN table_name VARCHAR(100),
ADD COLUMN allowed_fields TEXT;

-- Step 2: Migrate data from resource_name to table_name
UPDATE permissions
SET table_name = resource_name;

-- Step 3: Drop old columns
ALTER TABLE permissions
DROP COLUMN resource_name,
DROP COLUMN resource_type,
DROP COLUMN permission_action;

-- Step 4: Add constraints
ALTER TABLE permissions
ALTER COLUMN table_name SET NOT NULL;

-- Add comment for new columns
COMMENT ON COLUMN permissions.table_name IS '表名';
COMMENT ON COLUMN permissions.allowed_fields IS '允许查询的字段列表（JSON数组格式）';