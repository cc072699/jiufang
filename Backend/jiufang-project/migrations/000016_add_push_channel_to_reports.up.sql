ALTER TABLE scheduled_reports ADD COLUMN IF NOT EXISTS push_channel VARCHAR(20) NOT NULL DEFAULT 'wechat';
