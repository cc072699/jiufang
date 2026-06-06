-- 创建查询会话表
CREATE TABLE query_sessions (
    id BIGSERIAL PRIMARY KEY,
    snowflake_id BIGINT NOT NULL UNIQUE,
    user_id BIGINT NOT NULL,
    dialog_id BIGINT DEFAULT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    closed_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX idx_query_sessions_user_id ON query_sessions(user_id);
CREATE INDEX idx_query_sessions_status ON query_sessions(status);

COMMENT ON TABLE query_sessions IS '查询会话表';
COMMENT ON COLUMN query_sessions.snowflake_id IS '业务唯一标识(雪花ID)';
COMMENT ON COLUMN query_sessions.user_id IS '用户ID';
COMMENT ON COLUMN query_sessions.dialog_id IS '对话会话ID';
COMMENT ON COLUMN query_sessions.status IS '会话状态(active/closed)';

-- 创建查询记录表
CREATE TABLE query_records (
    id BIGSERIAL PRIMARY KEY,
    snowflake_id BIGINT NOT NULL UNIQUE,
    session_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    input TEXT NOT NULL,
    sql TEXT NOT NULL,
    status VARCHAR(20) NOT NULL,
    error_message TEXT DEFAULT NULL,
    result_count INTEGER DEFAULT NULL,
    execution_time INTEGER DEFAULT NULL,
    result_data TEXT DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_query_records_session_id ON query_records(session_id);
CREATE INDEX idx_query_records_user_id ON query_records(user_id);
CREATE INDEX idx_query_records_created_at ON query_records(created_at);

COMMENT ON TABLE query_records IS '查询记录表';
COMMENT ON COLUMN query_records.snowflake_id IS '业务唯一标识(雪花ID)';
COMMENT ON COLUMN query_records.session_id IS '查询会话ID';
COMMENT ON COLUMN query_records.user_id IS '用户ID';
COMMENT ON COLUMN query_records.input IS '用户输入的自然语言问题';
COMMENT ON COLUMN query_records.sql IS '生成的SQL语句';
COMMENT ON COLUMN query_records.status IS '查询状态(success/failed)';
COMMENT ON COLUMN query_records.error_message IS '错误信息(失败时)';
COMMENT ON COLUMN query_records.result_count IS '结果数量(成功时)';
COMMENT ON COLUMN query_records.execution_time IS '执行时间(毫秒)';
COMMENT ON COLUMN query_records.result_data IS '查询结果(JSON格式)';

-- 创建收藏表
CREATE TABLE favorites (
    id BIGSERIAL PRIMARY KEY,
    snowflake_id BIGINT NOT NULL UNIQUE,
    user_id BIGINT NOT NULL,
    query_record_id BIGINT NOT NULL,
    title VARCHAR(100) DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_favorites_user_id ON favorites(user_id);
CREATE INDEX idx_favorites_query_record_id ON favorites(query_record_id);
CREATE UNIQUE INDEX idx_favorites_unique ON favorites(user_id, query_record_id);

COMMENT ON TABLE favorites IS '收藏表';
COMMENT ON COLUMN favorites.snowflake_id IS '业务唯一标识(雪花ID)';
COMMENT ON COLUMN favorites.user_id IS '用户ID';
COMMENT ON COLUMN favorites.query_record_id IS '查询记录ID';
COMMENT ON COLUMN favorites.title IS '收藏标题(用户自定义)';