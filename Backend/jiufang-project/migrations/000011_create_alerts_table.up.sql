-- 创建预警规则表
CREATE TABLE alerts (
    id BIGSERIAL PRIMARY KEY,
    snowflake_id BIGINT NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(200) DEFAULT NULL,
    sql TEXT NOT NULL,
    condition VARCHAR(200) NOT NULL,
    recipients TEXT NOT NULL,
    push_channel VARCHAR(20) NOT NULL DEFAULT 'wechat',
    trigger_frequency VARCHAR(20) NOT NULL DEFAULT 'every_time',
    silence_start TIME DEFAULT NULL,
    silence_end TIME DEFAULT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    last_triggered_at TIMESTAMP DEFAULT NULL,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE UNIQUE INDEX idx_alerts_snowflake_id ON alerts(snowflake_id);
CREATE INDEX idx_alerts_created_by ON alerts(created_by);
CREATE INDEX idx_alerts_status ON alerts(status);
CREATE INDEX idx_alerts_last_triggered_at ON alerts(last_triggered_at);
CREATE INDEX idx_alerts_created_at ON alerts(created_at);

-- 创建更新时间触发器函数（如果不存在）
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 为alerts表创建触发器
CREATE TRIGGER update_alerts_updated_at
    BEFORE UPDATE ON alerts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 添加表注释
COMMENT ON TABLE alerts IS '预警规则表';
COMMENT ON COLUMN alerts.id IS '主键ID';
COMMENT ON COLUMN alerts.snowflake_id IS '雪花ID（业务唯一标识）';
COMMENT ON COLUMN alerts.name IS '预警规则名称';
COMMENT ON COLUMN alerts.description IS '预警规则描述';
COMMENT ON COLUMN alerts.sql IS '查询指标值的SQL语句';
COMMENT ON COLUMN alerts.condition IS '预警条件表达式（如 < 100）';
COMMENT ON COLUMN alerts.recipients IS '接收者列表（JSON数组）';
COMMENT ON COLUMN alerts.push_channel IS '推送渠道：wechat-企业微信，email-邮件';
COMMENT ON COLUMN alerts.trigger_frequency IS '触发频率：every_time-每次触发，daily-每日一次，weekly-每周一次';
COMMENT ON COLUMN alerts.silence_start IS '静默时段开始时间';
COMMENT ON COLUMN alerts.silence_end IS '静默时段结束时间';
COMMENT ON COLUMN alerts.status IS '预警规则状态：active-启用，inactive-停用';
COMMENT ON COLUMN alerts.last_triggered_at IS '上次触发时间';
COMMENT ON COLUMN alerts.created_by IS '创建者ID';
COMMENT ON COLUMN alerts.created_at IS '创建时间';
COMMENT ON COLUMN alerts.updated_at IS '更新时间';