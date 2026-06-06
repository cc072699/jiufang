-- 创建定时报告表
CREATE TABLE scheduled_reports (
    id BIGSERIAL PRIMARY KEY,
    snowflake_id BIGINT NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(200) DEFAULT NULL,
    query_condition TEXT NOT NULL,
    schedule_cron VARCHAR(50) NOT NULL,
    push_targets TEXT NOT NULL,
    push_channel VARCHAR(20) NOT NULL DEFAULT 'wechat',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_scheduled_reports_status ON scheduled_reports(status);
CREATE INDEX idx_scheduled_reports_created_by ON scheduled_reports(created_by);

-- 添加表注释
COMMENT ON TABLE scheduled_reports IS '定时报告表';
COMMENT ON COLUMN scheduled_reports.id IS '物理主键(自增ID)';
COMMENT ON COLUMN scheduled_reports.snowflake_id IS '业务唯一标识(雪花ID)';
COMMENT ON COLUMN scheduled_reports.name IS '报告名称';
COMMENT ON COLUMN scheduled_reports.description IS '报告描述';
COMMENT ON COLUMN scheduled_reports.query_condition IS '查询条件(JSON格式)';
COMMENT ON COLUMN scheduled_reports.schedule_cron IS '定时表达式(Cron格式)';
COMMENT ON COLUMN scheduled_reports.push_targets IS '推送对象(JSON数组,用户ID列表)';
COMMENT ON COLUMN scheduled_reports.push_channel IS '推送渠道';
COMMENT ON COLUMN scheduled_reports.status IS '报告状态';
COMMENT ON COLUMN scheduled_reports.created_by IS '创建者ID(外键)';
COMMENT ON COLUMN scheduled_reports.created_at IS '创建时间';
COMMENT ON COLUMN scheduled_reports.updated_at IS '更新时间';

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 为scheduled_reports表创建触发器
CREATE TRIGGER update_scheduled_reports_updated_at
    BEFORE UPDATE ON scheduled_reports
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 创建推送记录表
CREATE TABLE push_records (
    id BIGSERIAL PRIMARY KEY,
    snowflake_id BIGINT NOT NULL UNIQUE,
    report_id BIGINT DEFAULT NULL,
    alert_rule_id BIGINT DEFAULT NULL,
    push_type VARCHAR(20) NOT NULL,
    push_content TEXT NOT NULL,
    push_targets TEXT NOT NULL,
    push_channel VARCHAR(20) NOT NULL,
    push_status VARCHAR(20) NOT NULL,
    push_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    error_message VARCHAR(500) DEFAULT NULL,
    retry_count INTEGER NOT NULL DEFAULT 0
);

-- 创建索引
CREATE INDEX idx_push_records_report_id ON push_records(report_id);
CREATE INDEX idx_push_records_alert_rule_id ON push_records(alert_rule_id);
CREATE INDEX idx_push_records_push_time ON push_records(push_time);

-- 添加表注释
COMMENT ON TABLE push_records IS '推送记录表';
COMMENT ON COLUMN push_records.id IS '物理主键(自增ID)';
COMMENT ON COLUMN push_records.snowflake_id IS '业务唯一标识(雪花ID)';
COMMENT ON COLUMN push_records.report_id IS '定时报告ID(外键)';
COMMENT ON COLUMN push_records.alert_rule_id IS '预警规则ID(外键)';
COMMENT ON COLUMN push_records.push_type IS '推送类型';
COMMENT ON COLUMN push_records.push_content IS '推送内容(Markdown格式)';
COMMENT ON COLUMN push_records.push_targets IS '推送对象(JSON数组)';
COMMENT ON COLUMN push_records.push_channel IS '推送渠道';
COMMENT ON COLUMN push_records.push_status IS '推送状态';
COMMENT ON COLUMN push_records.push_time IS '推送时间';
COMMENT ON COLUMN push_records.error_message IS '错误消息';
COMMENT ON COLUMN push_records.retry_count IS '重试次数';