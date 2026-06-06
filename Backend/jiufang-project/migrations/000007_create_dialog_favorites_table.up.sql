-- 创建对话收藏表
-- Migration: 000007_create_dialog_favorites_table
-- Description: 创建对话收藏表，用于存储用户收藏的对话会话
-- Author: AI Agent
-- Date: 2026-06-03

CREATE TABLE dialog_favorites (
    id BIGSERIAL PRIMARY KEY,
    snowflake_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    dialog_session_id BIGINT NOT NULL,
    title VARCHAR(100) DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (snowflake_id),
    UNIQUE (user_id, dialog_session_id)
);

-- 创建索引
CREATE INDEX idx_dialog_favorites_user_id ON dialog_favorites(user_id);
CREATE INDEX idx_dialog_favorites_dialog_session_id ON dialog_favorites(dialog_session_id);

-- 添加表注释
COMMENT ON TABLE dialog_favorites IS '对话收藏表';

-- 添加字段注释
COMMENT ON COLUMN dialog_favorites.id IS '物理主键(自增ID)';
COMMENT ON COLUMN dialog_favorites.snowflake_id IS '业务唯一标识(雪花ID)';
COMMENT ON COLUMN dialog_favorites.user_id IS '用户ID';
COMMENT ON COLUMN dialog_favorites.dialog_session_id IS '对话会话ID';
COMMENT ON COLUMN dialog_favorites.title IS '收藏标题(用户自定义)';
COMMENT ON COLUMN dialog_favorites.created_at IS '创建时间';