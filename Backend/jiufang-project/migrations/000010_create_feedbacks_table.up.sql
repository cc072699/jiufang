-- 创建反馈表
CREATE TABLE feedbacks (
    id BIGSERIAL PRIMARY KEY,
    snowflake_id BIGINT NOT NULL UNIQUE,
    user_id BIGINT NOT NULL,
    query_record_id BIGINT NOT NULL,
    query_question TEXT NOT NULL,
    rating VARCHAR(20) NOT NULL CHECK (rating IN ('satisfied', 'unsatisfied')),
    reason TEXT DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE UNIQUE INDEX idx_feedbacks_snowflake_id ON feedbacks(snowflake_id);
CREATE INDEX idx_feedbacks_user_id ON feedbacks(user_id);
CREATE INDEX idx_feedbacks_query_record_id ON feedbacks(query_record_id);
CREATE INDEX idx_feedbacks_rating ON feedbacks(rating);
CREATE INDEX idx_feedbacks_created_at ON feedbacks(created_at);

-- 添加表注释
COMMENT ON TABLE feedbacks IS '用户反馈表';
COMMENT ON COLUMN feedbacks.id IS '主键ID';
COMMENT ON COLUMN feedbacks.snowflake_id IS '雪花ID（业务唯一标识）';
COMMENT ON COLUMN feedbacks.user_id IS '用户ID';
COMMENT ON COLUMN feedbacks.query_record_id IS '查询记录ID';
COMMENT ON COLUMN feedbacks.query_question IS '查询问句';
COMMENT ON COLUMN feedbacks.rating IS '评价结果：satisfied-满意，unsatisfied-不满意';
COMMENT ON COLUMN feedbacks.reason IS '不满意原因（rating为unsatisfied时必填）';
COMMENT ON COLUMN feedbacks.created_at IS '创建时间';