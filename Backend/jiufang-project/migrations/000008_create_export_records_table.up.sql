-- 创建导出记录表
-- Migration: 000008_create_export_records_table
-- Description: 创建导出记录表，用于存储用户的查询结果导出历史
-- Author: AI Agent
-- Date: 2026-06-03

CREATE TABLE export_records (
    id BIGSERIAL PRIMARY KEY,
    snowflake_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    query_record_id BIGINT NOT NULL,
    format VARCHAR(20) NOT NULL,
    file_name VARCHAR(200) NOT NULL,
    file_size BIGINT NOT NULL,
    query_summary TEXT DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (snowflake_id)
);

-- 创建索引
CREATE INDEX idx_export_records_user_id ON export_records(user_id);
CREATE INDEX idx_export_records_created_at ON export_records(created_at);

-- 添加表注释
COMMENT ON TABLE export_records IS '导出记录表';

-- 添加字段注释
COMMENT ON COLUMN export_records.id IS '物理主键(自增ID)';
COMMENT ON COLUMN export_records.snowflake_id IS '业务唯一标识(雪花ID)';
COMMENT ON COLUMN export_records.user_id IS '用户ID';
COMMENT ON COLUMN export_records.query_record_id IS '查询记录ID';
COMMENT ON COLUMN export_records.format IS '导出格式(excel/pdf)';
COMMENT ON COLUMN export_records.file_name IS '导出文件名';
COMMENT ON COLUMN export_records.file_size IS '文件大小(字节)';
COMMENT ON COLUMN export_records.query_summary IS '查询条件摘要';
COMMENT ON COLUMN export_records.created_at IS '导出时间';