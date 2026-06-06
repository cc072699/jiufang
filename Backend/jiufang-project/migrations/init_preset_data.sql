-- =====================================================
-- ERP对话式查询助手 - 预设数据初始化脚本
-- =====================================================
-- 版本：1.0
-- 日期：2026-06-05
-- 说明：本脚本用于初始化PRD文档要求的预设数据
-- =====================================================

-- =====================================================
-- 1. 用户组预置数据（PRD BR-032要求）
-- =====================================================
-- 注意：用户组预置数据已在000003_create_permission_tables.up.sql中定义
-- 如果该迁移脚本未执行，请先执行迁移脚本
-- 如果已执行，以下INSERT语句会因唯一约束冲突而失败，可以跳过

-- 检查用户组是否已存在，如果不存在则插入
INSERT INTO user_groups (snowflake_id, name, description, created_at, updated_at)
SELECT 1000000000000000001, '管理员组', '系统管理员，拥有全部权限', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM user_groups WHERE snowflake_id = 1000000000000000001);

INSERT INTO user_groups (snowflake_id, name, description, created_at, updated_at)
SELECT 1000000000000000002, '高管组', '公司高管，可查询全公司数据', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM user_groups WHERE snowflake_id = 1000000000000000002);

INSERT INTO user_groups (snowflake_id, name, description, created_at, updated_at)
SELECT 1000000000000000003, '采购组', '采购部门，可查询采购相关数据', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM user_groups WHERE snowflake_id = 1000000000000000003);

INSERT INTO user_groups (snowflake_id, name, description, created_at, updated_at)
SELECT 1000000000000000004, '销售组', '销售部门，可查询销售相关数据', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM user_groups WHERE snowflake_id = 1000000000000000004);

INSERT INTO user_groups (snowflake_id, name, description, created_at, updated_at)
SELECT 1000000000000000005, '财务组', '财务部门，可查询财务相关数据', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM user_groups WHERE snowflake_id = 1000000000000000005);

-- =====================================================
-- 2. 权限配置数据（示例配置）
-- =====================================================
-- 说明：根据PRD文档的权限矩阵配置各用户组的权限
-- 注意：以下权限配置为示例配置，实际配置应根据ERP数据库表结构调整

-- 2.1 管理员组权限配置
-- 管理员组拥有全部权限，可以查询所有表的所有字段
INSERT INTO permissions (snowflake_id, group_id, table_name, allowed_fields, filter_condition, created_at, updated_at)
SELECT 2000000000000000001, 
       (SELECT id FROM user_groups WHERE snowflake_id = 1000000000000000001),
       'users', 
       '["id", "username", "email", "role", "status", "created_at", "updated_at"]',
       NULL,
       CURRENT_TIMESTAMP, 
       CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE snowflake_id = 2000000000000000001);

INSERT INTO permissions (snowflake_id, group_id, table_name, allowed_fields, filter_condition, created_at, updated_at)
SELECT 2000000000000000002, 
       (SELECT id FROM user_groups WHERE snowflake_id = 1000000000000000001),
       'user_groups', 
       '["id", "name", "description", "created_at", "updated_at"]',
       NULL,
       CURRENT_TIMESTAMP, 
       CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE snowflake_id = 2000000000000000002);

INSERT INTO permissions (snowflake_id, group_id, table_name, allowed_fields, filter_condition, created_at, updated_at)
SELECT 2000000000000000003, 
       (SELECT id FROM user_groups WHERE snowflake_id = 1000000000000000001),
       'permissions', 
       '["id", "group_id", "table_name", "allowed_fields", "filter_condition", "created_at", "updated_at"]',
       NULL,
       CURRENT_TIMESTAMP, 
       CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE snowflake_id = 2000000000000000003);

INSERT INTO permissions (snowflake_id, group_id, table_name, allowed_fields, filter_condition, created_at, updated_at)
SELECT 2000000000000000004, 
       (SELECT id FROM user_groups WHERE snowflake_id = 1000000000000000001),
       'operation_logs', 
       '["id", "user_id", "operation_type", "operation_detail", "created_at"]',
       NULL,
       CURRENT_TIMESTAMP, 
       CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE snowflake_id = 2000000000000000004);

-- 2.2 高管组权限配置
-- 高管组可以查询全公司数据，包括敏感字段（成本价、利润率）
INSERT INTO permissions (snowflake_id, group_id, table_name, allowed_fields, filter_condition, created_at, updated_at)
SELECT 2000000000000000005, 
       (SELECT id FROM user_groups WHERE snowflake_id = 1000000000000000002),
       'users', 
       '["id", "username", "email", "role", "status", "created_at", "updated_at"]',
       NULL,
       CURRENT_TIMESTAMP, 
       CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE snowflake_id = 2000000000000000005);

-- 注意：以下ERP表的权限配置需要根据实际ERP数据库表结构调整
-- 示例：假设ERP数据库有以下表
-- purchase_orders（采购订单表）
-- sales_orders（销售订单表）
-- inventory（库存表）
-- accounts_payable（应付款表）
-- accounts_receivable（应收款表）

INSERT INTO permissions (snowflake_id, group_id, table_name, allowed_fields, filter_condition, created_at, updated_at)
SELECT 2000000000000000006, 
       (SELECT id FROM user_groups WHERE snowflake_id = 1000000000000000002),
       'purchase_orders', 
       '["id", "order_no", "supplier_name", "total_amount", "cost_price", "profit_rate", "order_date", "status"]',
       NULL,
       CURRENT_TIMESTAMP, 
       CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE snowflake_id = 2000000000000000006);

INSERT INTO permissions (snowflake_id, group_id, table_name, allowed_fields, filter_condition, created_at, updated_at)
SELECT 2000000000000000007, 
       (SELECT id FROM user_groups WHERE snowflake_id = 1000000000000000002),
       'sales_orders', 
       '["id", "order_no", "customer_name", "total_amount", "cost_price", "profit_rate", "order_date", "status"]',
       NULL,
       CURRENT_TIMESTAMP, 
       CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE snowflake_id = 2000000000000000007);

INSERT INTO permissions (snowflake_id, group_id, table_name, allowed_fields, filter_condition, created_at, updated_at)
SELECT 2000000000000000008, 
       (SELECT id FROM user_groups WHERE snowflake_id = 1000000000000000002),
       'inventory', 
       '["id", "product_name", "quantity", "cost_price", "profit_rate", "warehouse", "last_updated"]',
       NULL,
       CURRENT_TIMESTAMP, 
       CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE snowflake_id = 2000000000000000008);

-- 2.3 采购组权限配置
-- 采购组只能查询采购相关数据，不能查看敏感字段（成本价、利润率）
INSERT INTO permissions (snowflake_id, group_id, table_name, allowed_fields, filter_condition, created_at, updated_at)
SELECT 2000000000000000009, 
       (SELECT id FROM user_groups WHERE snowflake_id = 1000000000000000003),
       'purchase_orders', 
       '["id", "order_no", "supplier_name", "total_amount", "order_date", "status"]',
       'department = "采购部"', -- 数据级权限过滤条件
       CURRENT_TIMESTAMP, 
       CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE snowflake_id = 2000000000000000009);

-- 2.4 销售组权限配置
-- 销售组只能查询销售相关数据，不能查看敏感字段（成本价、利润率）
INSERT INTO permissions (snowflake_id, group_id, table_name, allowed_fields, filter_condition, created_at, updated_at)
SELECT 2000000000000000010, 
       (SELECT id FROM user_groups WHERE snowflake_id = 1000000000000000004),
       'sales_orders', 
       '["id", "order_no", "customer_name", "total_amount", "order_date", "status"]',
       'department = "销售部"', -- 数据级权限过滤条件
       CURRENT_TIMESTAMP, 
       CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE snowflake_id = 2000000000000000010);

-- 2.5 财务组权限配置
-- 财务组可以查询财务相关数据，可以查看敏感字段（成本价、利润率）
INSERT INTO permissions (snowflake_id, group_id, table_name, allowed_fields, filter_condition, created_at, updated_at)
SELECT 2000000000000000011, 
       (SELECT id FROM user_groups WHERE snowflake_id = 1000000000000000005),
       'accounts_payable', 
       '["id", "supplier_name", "amount", "due_date", "status", "cost_price", "profit_rate"]',
       NULL,
       CURRENT_TIMESTAMP, 
       CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE snowflake_id = 2000000000000000011);

INSERT INTO permissions (snowflake_id, group_id, table_name, allowed_fields, filter_condition, created_at, updated_at)
SELECT 2000000000000000012, 
       (SELECT id FROM user_groups WHERE snowflake_id = 1000000000000000005),
       'accounts_receivable', 
       '["id", "customer_name", "amount", "due_date", "status", "cost_price", "profit_rate"]',
       NULL,
       CURRENT_TIMESTAMP, 
       CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE snowflake_id = 2000000000000000012);

-- =====================================================
-- 3. 预警规则预置数据（PRD BR-044要求）
-- =====================================================
-- 系统预置常用预警规则：库存低于安全库存、应付款账期超期、采购单价异常波动

-- 3.1 库存低于安全库存预警
INSERT INTO alerts (snowflake_id, name, description, sql, condition, recipients, push_channel, trigger_frequency, status, created_by, created_at, updated_at)
SELECT 3000000000000000001, 
       '库存低于安全库存预警', 
       '当库存数量低于安全库存阈值时触发预警',
       'SELECT product_name, quantity, safety_stock FROM inventory WHERE quantity < safety_stock',
       'quantity < safety_stock',
       '[1001, 1002]', -- 接收者列表（用户ID数组）
       'wechat',
       'every_time',
       'active',
       1, -- 创建者ID（管理员）
       CURRENT_TIMESTAMP, 
       CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM alerts WHERE snowflake_id = 3000000000000000001);

-- 3.2 应付款账期超期预警
INSERT INTO alerts (snowflake_id, name, description, sql, condition, recipients, push_channel, trigger_frequency, status, created_by, created_at, updated_at)
SELECT 3000000000000000002, 
       '应付款账期超期预警', 
       '当应付款账期超过约定天数时触发预警',
       'SELECT supplier_name, amount, due_date, DATEDIFF(NOW(), due_date) as overdue_days FROM accounts_payable WHERE due_date < NOW() AND status != "已付款"',
       'overdue_days > 0',
       '[1001, 1002, 1003]', -- 接收者列表（用户ID数组）
       'wechat',
       'daily',
       'active',
       1, -- 创建者ID（管理员）
       CURRENT_TIMESTAMP, 
       CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM alerts WHERE snowflake_id = 3000000000000000002);

-- 3.3 采购单价异常波动预警
INSERT INTO alerts (snowflake_id, name, description, sql, condition, recipients, push_channel, trigger_frequency, status, created_by, created_at, updated_at)
SELECT 3000000000000000003, 
       '采购单价异常波动预警', 
       '当采购单价波动超过20%时触发预警',
       'SELECT product_name, unit_price, avg_price, ABS(unit_price - avg_price) / avg_price * 100 as fluctuation_rate FROM (SELECT product_name, unit_price, AVG(unit_price) OVER (PARTITION BY product_name ORDER BY order_date ROWS BETWEEN 30 PRECEDING AND CURRENT ROW) as avg_price FROM purchase_orders) WHERE ABS(unit_price - avg_price) / avg_price * 100 > 20',
       'fluctuation_rate > 20',
       '[1001, 1002]', -- 接收者列表（用户ID数组）
       'wechat',
       'every_time',
       'active',
       1, -- 创建者ID（管理员）
       CURRENT_TIMESTAMP, 
       CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM alerts WHERE snowflake_id = 3000000000000000003);

-- =====================================================
-- 4. 管理员账号初始化
-- =====================================================
-- 插入两个拥有全部权限的管理员账号
-- 用户名：admin1、admin2
-- 密码：214510115lhl（已使用bcrypt加密）
-- 角色：admin（拥有全部权限）

INSERT INTO users (snowflake_id, username, password, email, role, status, is_first_login, created_at, updated_at)
SELECT 1780644326008, 
       'admin1', 
       '$2a$10$HkaIu1QU5P01N8NRJxOT5ep8nGLbh59qh.kd4AD9OmiZzxmj57ylu', -- bcrypt加密密码
       'admin1@jiufang.com', 
       'admin', 
       1, 
       false, 
       CURRENT_TIMESTAMP, 
       CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM users WHERE snowflake_id = 1780644326008);

INSERT INTO users (snowflake_id, username, password, email, role, status, is_first_login, created_at, updated_at)
SELECT 1780644326009, 
       'admin2', 
       '$2a$10$HkaIu1QU5P01N8NRJxOT5ep8nGLbh59qh.kd4AD9OmiZzxmj57ylu', -- bcrypt加密密码
       'admin2@jiufang.com', 
       'admin', 
       1, 
       false, 
       CURRENT_TIMESTAMP, 
       CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM users WHERE snowflake_id = 1780644326009);

-- =====================================================
-- 5. 用户组成员关联初始化
-- =====================================================
-- 将管理员账号添加到管理员组

INSERT INTO user_group_members (snowflake_id, user_id, group_id, created_at)
SELECT 4000000000000000001, 
       (SELECT id FROM users WHERE snowflake_id = 1780644326008),
       (SELECT id FROM user_groups WHERE snowflake_id = 1000000000000000001),
       CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM user_group_members WHERE snowflake_id = 4000000000000000001);

INSERT INTO user_group_members (snowflake_id, user_id, group_id, created_at)
SELECT 4000000000000000002, 
       (SELECT id FROM users WHERE snowflake_id = 1780644326009),
       (SELECT id FROM user_groups WHERE snowflake_id = 1000000000000000001),
       CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM user_group_members WHERE snowflake_id = 4000000000000000002);

-- =====================================================
-- 6. 验证初始化结果
-- =====================================================
-- 执行以下查询验证数据是否正确初始化

-- 6.1 验证用户组
SELECT '用户组初始化验证' as check_type, 
       COUNT(*) as expected_count, 
       (SELECT COUNT(*) FROM user_groups WHERE snowflake_id IN (1000000000000000001, 1000000000000000002, 1000000000000000003, 1000000000000000004, 1000000000000000005)) as actual_count,
       CASE WHEN COUNT(*) = (SELECT COUNT(*) FROM user_groups WHERE snowflake_id IN (1000000000000000001, 1000000000000000002, 1000000000000000003, 1000000000000000004, 1000000000000000005)) THEN '✅ 成功' ELSE '❌ 失败' END as status
FROM (SELECT 1) as temp WHERE 5 = 5;

-- 6.2 验证权限配置
SELECT '权限配置初始化验证' as check_type, 
       COUNT(*) as expected_count, 
       (SELECT COUNT(*) FROM permissions WHERE snowflake_id BETWEEN 2000000000000000001 AND 2000000000000000012) as actual_count,
       CASE WHEN COUNT(*) = (SELECT COUNT(*) FROM permissions WHERE snowflake_id BETWEEN 2000000000000000001 AND 2000000000000000012) THEN '✅ 成功' ELSE '❌ 失败' END as status
FROM (SELECT 1) as temp WHERE 12 = 12;

-- 6.3 验证预警规则
SELECT '预警规则初始化验证' as check_type, 
       COUNT(*) as expected_count, 
       (SELECT COUNT(*) FROM alerts WHERE snowflake_id IN (3000000000000000001, 3000000000000000002, 3000000000000000003)) as actual_count,
       CASE WHEN COUNT(*) = (SELECT COUNT(*) FROM alerts WHERE snowflake_id IN (3000000000000000001, 3000000000000000002, 3000000000000000003)) THEN '✅ 成功' ELSE '❌ 失败' END as status
FROM (SELECT 1) as temp WHERE 3 = 3;

-- 6.4 验证管理员账号
SELECT '管理员账号初始化验证' as check_type, 
       COUNT(*) as expected_count, 
       (SELECT COUNT(*) FROM users WHERE snowflake_id IN (1780644326008, 1780644326009)) as actual_count,
       CASE WHEN COUNT(*) = (SELECT COUNT(*) FROM users WHERE snowflake_id IN (1780644326008, 1780644326009)) THEN '✅ 成功' ELSE '❌ 失败' END as status
FROM (SELECT 1) as temp WHERE 2 = 2;

-- 6.5 验证用户组成员关联
SELECT '用户组成员关联初始化验证' as check_type, 
       COUNT(*) as expected_count, 
       (SELECT COUNT(*) FROM user_group_members WHERE snowflake_id IN (4000000000000000001, 4000000000000000002)) as actual_count,
       CASE WHEN COUNT(*) = (SELECT COUNT(*) FROM user_group_members WHERE snowflake_id IN (4000000000000000001, 4000000000000000002)) THEN '✅ 成功' ELSE '❌ 失败' END as status
FROM (SELECT 1) as temp WHERE 2 = 2;

-- =====================================================
-- 初始化完成提示
-- =====================================================
SELECT '========================================' as message;
SELECT '预设数据初始化完成！' as message;
SELECT '========================================' as message;
SELECT '初始化内容：' as message;
SELECT '1. 用户组预置数据（5个用户组）' as message;
SELECT '2. 权限配置数据（12条权限记录）' as message;
SELECT '3. 预警规则预置数据（3条预警规则）' as message;
SELECT '4. 管理员账号（2个管理员账号）' as message;
SELECT '5. 用户组成员关联（2条关联记录）' as message;
SELECT '========================================' as message;
SELECT '管理员账号信息：' as message;
SELECT '用户名：admin1 / admin2' as message;
SELECT '密码：214510115lhl' as message;
SELECT '角色：admin（拥有全部权限）' as message;
SELECT '========================================' as message;
SELECT '请使用管理员账号登录系统进行后续配置' as message;
SELECT '========================================' as message;