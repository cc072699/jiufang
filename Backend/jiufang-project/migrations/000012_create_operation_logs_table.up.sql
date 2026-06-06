-- 创建操作日志表
CREATE TABLE operation_logs (
    id BIGSERIAL PRIMARY KEY,
    snowflake_id BIGINT NOT NULL UNIQUE,
    user_id BIGINT DEFAULT NULL,
    operation_type VARCHAR(50) NOT NULL,
    operation_object VARCHAR(100) DEFAULT NULL,
    operation_detail TEXT DEFAULT NULL,
    operation_result VARCHAR(20) NOT NULL CHECK (operation_result IN ('success', 'failed')),
    ip_address VARCHAR(50) DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE UNIQUE INDEX idx_operation_logs_snowflake_id ON operation_logs(snowflake_id);
CREATE INDEX idx_operation_logs_user_id ON operation_logs(user_id);
CREATE INDEX idx_operation_logs_operation_type ON operation_logs(operation_type);
CREATE INDEX idx_operation_logs_created_at ON operation_logs(created_at);

-- 添加表注释
COMMENT ON TABLE operation_logs IS '操作日志表';
COMMENT ON COLUMN operation_logs.id IS '主键ID';
COMMENT ON COLUMN operation_logs.snowflake_id IS '雪花ID（业务唯一标识）';
COMMENT ON COLUMN operation_logs.user_id IS '操作者ID（关联users表）';
COMMENT ON COLUMN operation_logs.operation_type IS '操作类型（login/logout/query/create_user/update_user/delete_user/config_permission/create_report等）';
COMMENT ON COLUMN operation_logs.operation_object IS '操作对象（如用户名、报告名）';
COMMENT ON COLUMN operation_logs.operation_detail IS '操作详情（JSON格式）';
COMMENT ON COLUMN operation_logs.operation_result IS '操作结果（success/failed）';
COMMENT ON COLUMN operation_logs.ip_address IS '操作者IP地址';
COMMENT ON COLUMN operation_logs.created_at IS '操作时间';