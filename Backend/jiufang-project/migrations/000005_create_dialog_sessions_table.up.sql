-- 创建对话会话表
-- Migration: 000005_create_dialog_sessions_table
-- Description: 创建对话会话表，用于存储多轮对话的会话元数据
-- Author: AI Agent
-- Date: 2026-06-03

CREATE TABLE dialog_sessions (
    id BIGSERIAL PRIMARY KEY,
    snowflake_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    query_session_id BIGINT DEFAULT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    closed_at TIMESTAMP DEFAULT NULL,
    UNIQUE (snowflake_id)
);

-- 创建索引
CREATE INDEX idx_dialog_sessions_user_id ON dialog_sessions(user_id);
CREATE INDEX idx_dialog_sessions_status ON dialog_sessions(status);

-- 添加表注释
COMMENT ON TABLE dialog_sessions IS '对话会话表';

-- 添加字段注释
COMMENT ON COLUMN dialog_sessions.id IS '物理主键(自增ID)';
COMMENT ON COLUMN dialog_sessions.snowflake_id IS '业务唯一标识(雪花ID)';
COMMENT ON COLUMN dialog_sessions.user_id IS '用户ID(外键)';
COMMENT ON COLUMN dialog_sessions.query_session_id IS '关联的查询会话ID';
COMMENT ON COLUMN dialog_sessions.status IS '会话状态(active/closed)';
COMMENT ON COLUMN dialog_sessions.created_at IS '创建时间';
COMMENT ON COLUMN dialog_sessions.updated_at IS '更新时间';
COMMENT ON COLUMN dialog_sessions.closed_at IS '关闭时间';

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 创建触发器
CREATE TRIGGER update_dialog_sessions_updated_at 
    BEFORE UPDATE ON dialog_sessions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();