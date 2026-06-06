-- Rollback: Restore permissions table to original structure

-- Step 1: Add back old columns
ALTER TABLE permissions
ADD COLUMN resource_name VARCHAR(100),
ADD COLUMN resource_type VARCHAR(50),
ADD COLUMN permission_action VARCHAR(50);

-- Step 2: Migrate data back
UPDATE permissions
SET resource_name = table_name;

-- Step 3: Set default values for removed columns
UPDATE permissions
SET resource_type = 'table',
    permission_action = 'query';

-- Step 4: Drop new columns
ALTER TABLE permissions
DROP COLUMN table_name,
DROP COLUMN allowed_fields;

-- Step 5: Restore constraints
ALTER TABLE permissions
ALTER COLUMN resource_name SET NOT NULL;

-- Remove comments
COMMENT ON COLUMN permissions.table_name IS NULL;
COMMENT ON COLUMN permissions.allowed_fields IS NULL;