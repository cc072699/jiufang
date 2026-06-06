---
name: integration-test-developer
description: 当用户提出以下需求时触发：为系统生成集成测试、编写模块间交互测试、搭建集成测试环境、配置 Testcontainers/Docker 测试基座、验证 API 端到端链路、测试数据库/缓存/消息队列集成、执行并运行集成测试套件。
---

# 集成测试开发与执行 Skill

## 核心原则
1. **文档先行，绝不绕过**：任何集成测试开发或执行前，必须先检查并读取相关项目文档（PRD、架构文档、API 契约、数据库设计、代码规范）。此步骤为强制步骤，绝对不允许跳过。
2. **方案先行，确认执行**：生成任何测试代码或测试配置前，必须先输出完整的《集成测试实施方案》，等待用户明确确认（如回复"确认"、"同意"、"执行"、"按方案来"）。未确认前严禁动手。
3. **阻断执行，未经确认不动手**：在用户明确确认方案之前，严禁生成测试代码、创建测试文件、修改现有代码或执行测试命令。此阻断机制无任何例外。
4. **生成与执行分离，二次确认**：集成测试代码生成完毕后，必须先询问用户是否需要执行测试。用户明确确认执行后，才能运行测试套件。严禁在生成阶段擅自执行测试。
5. **真实依赖，可控隔离**：集成测试允许使用真实的外部依赖（数据库、缓存、消息队列），但必须通过 Testcontainers、Docker Compose 或内嵌服务器实现环境隔离，确保测试可重复、不污染生产环境。
6. **链路验证，关注交互**：集成测试的核心是验证模块间的接口契约、数据流和交互行为，而非单个函数的内部逻辑。
7. **可自动化执行**：生成的集成测试必须能一键运行（如 `mvn test -Pintegration-test`、`pytest tests/integration/`、`docker-compose -f docker-compose.test.yml up --abort-on-container-exit`），并提供清晰的执行指南。

## 执行工作流

### 阶段一：文档研读与环境分析（自动执行，强制步骤）
当本技能被触发后，AI 必须首先执行以下步骤，不得跳过：

1. **读取项目文档**（使用文件工具读取，必须全部读取完成后才能进入下一步）：
   - **PRD / API 文档**：`docs/PRD.md`、`docs/api.md`、`openapi.yaml` 或用户指定的接口文档
   - **架构文档**：`docs/architecture.md`、`docs/技术架构概要.md`，了解模块划分和依赖关系
   - **数据库设计**：`docs/database.md`、`docs/数据库设计.md`、`migrations/`（数据库迁移脚本）
   - **Trae规则文件**：`.trae/rules/gin-base.md`（日志规范、监控规范）、`.trae/rules/gin-project.md` 或 `.trae/rules/gin-eino-project.md`（项目规范）
   - **现有测试配置**：`go.mod`、`Makefile`、`docker-compose.yml`、`docker-compose.test.yml` 等

   **文档读取规则**：
   - AI 必须依次尝试读取以上文档
   - 必须全部读取完成后，才能执行后续分析
   - 只要有任何一个关键文档未读取到，AI 必须立即询问用户，用户确认可跳过后才能跳过
   - 此规则适用于所有场景，无例外

2. **分析并提取关键信息**：
   - 被测链路：HTTP 入口 → Handler → Service → Repository → 数据库/缓存/MQ
   - 外部依赖清单：数据库类型版本、缓存（Redis）、消息队列（Kafka/RabbitMQ）、第三方 HTTP 服务
   - 接口契约：请求/响应格式、状态码定义、错误码规范
   - 数据初始化需求：测试前置数据（SQL 脚本、Fixture、Factory）
   - 环境基座：是否已有 Docker Compose、testcontainers 配置、CI/CD 测试流水线

3. **一致性检查**：
   - 检查 API 契约与代码实现是否一致
   - 检查数据库脚本版本与实体类是否匹配
   - 若发现冲突或缺失，在方案中标注风险点，不擅自假设

### 阶段二：方案制定（必须输出）
基于阶段一的分析，输出《集成测试实施方案》，格式如下：

```markdown
## 📋 集成测试实施方案

### 1. 测试目标
- 被测链路：
- 测试范围：□ API 端到端  □ 数据库集成  □ 缓存集成  □ 消息队列集成  □ 第三方服务集成
- 编程语言与测试框架：

### 2. 环境基座设计
- 基座类型：□ Testcontainers  □ Docker Compose  □ 内嵌服务器（H2/Embedded Redis）  □ 现有测试环境
- 依赖服务及版本：
  - 数据库：
  - 缓存：
  - 消息队列：
  - 其他：
- 端口映射与网络配置：

### 3. 目录与文件结构
```
project-root/
├── src/test/integration/ 或 tests/integration/
│   ├── base/
│   │   └── IntegrationTestBase.java / conftest.py
│   ├── fixtures/
│   │   ├── sql/
│   │   └── data_factory.py
│   ├── xxx_integration_test.java / test_xxx_integration.py
│   └── docker-compose.test.yml（如需要）
```

### 4. 测试用例清单
| 用例编号 | 测试场景 | 前置条件 | 操作步骤 | 预期结果 | 验证点 |
|---------|---------|---------|---------|---------|--------|
| IT-01 | 正常注册流程 | 数据库为空 | POST /api/users | 201 Created | 数据库存在记录、响应体正确 |
| IT-02 | 重复注册 | 已存在用户 | POST /api/users | 409 Conflict | 数据库记录数不变 |
| IT-03 | 数据库异常 | 模拟连接断开 | POST /api/users | 500 Error | 事务回滚、无脏数据 |

### 5. 数据初始化与清理策略
- 前置数据：SQL 脚本 / Fixture / Factory
- 清理策略：@Transactional 回滚 / @DirtiesContext / TRUNCATE 脚本 / 容器重启

### 6. 执行命令
- 本地执行：
- CI/CD 执行：

### 7. 关键设计决策与风险点
- 决策1：（说明原因）
- 风险1：（说明缓解措施）

### 8. 规范合规自检项
- [ ] 测试命名符合规范
- [ ] 环境隔离，不污染开发/生产数据
- [ ] 所有外部依赖通过容器/内嵌方式管理
- [ ] 测试数据清理策略明确
- [ ] 执行命令可一键运行
```

**输出后，AI 必须明确提示：**
> 以上是针对「XXX链路/模块」的集成测试实施方案。请审阅方案内容，确认无误后回复「确认执行」，我将严格按照本方案生成测试代码并配置环境。若需调整，请直接指出修改点。

### 阶段三：等待用户确认（第一阻断点，绝对不可绕过）
- **若用户未明确确认**：仅回答疑问、接受修改意见、更新方案，绝不创建任何文件、生成测试代码、修改现有代码或执行测试命令。
- **若用户明确确认**（如回复"确认"、"同意"、"执行"、"按方案来"）：进入阶段四。
- **此阻断机制适用于所有场景（生成集成测试、修改现有集成测试、追加集成测试场景），无任何例外。**

### 阶段四：代码生成与环境配置（确认后执行）
1. **环境基座搭建**：
   - 若使用 Testcontainers：生成容器配置类，定义数据库/Redis/Kafka 容器生命周期
   - 若使用 Docker Compose：生成 `docker-compose.test.yml`，配置服务、网络、卷映射
   - 若使用内嵌服务：配置 H2、Embedded Redis、Embedded Kafka 等
   - 每完成一个环境组件配置，简要汇报进度

2. **测试基座类生成**：
   - 生成所有集成测试共用的基座类（如 `IntegrationTestBase`），包含上下文加载、容器初始化、数据库连接工具、HTTP 客户端配置
   - 添加基于代码规范要求的注释头

3. **测试用例逐一生成**：
   - 按方案中的用例清单顺序生成测试代码
   - 每个测试文件顶部添加注释头（作者、日期、被测链路、功能描述）
   - 遵循 Arrange-Act-Assert（AAA）结构
   - 对真实依赖验证：数据库记录状态、缓存键值、消息队列消费情况、HTTP 响应码和响应体

4. **数据 Fixture / Factory 生成**：
   - 生成 SQL 初始化脚本或代码级 Factory，用于构造前置测试数据
   - 确保数据之间无冲突，支持并行测试

5. **实时自检**：
   - 每生成一段测试逻辑，对照方案检查是否遗漏验证点
   - 检查是否意外引入了对生产环境的依赖

6. **阶段四完成汇报**：
   - 输出《集成测试代码交付清单》，列出所有新建/修改的文件
   - 明确提示用户：
     > 集成测试代码及环境配置已全部生成完毕。请问是否需要现在执行测试套件？
     > 如需执行，请回复「执行测试」或「确认运行」，我将启动测试环境并运行所有集成测试。
     > 如暂不需要执行，可回复「不执行」或忽略此消息。

### 阶段五：等待用户确认是否执行（第二阻断点，绝对不可绕过）
- **若用户未明确要求执行或回复「不执行」**：不执行任何测试命令，仅回答疑问或接受进一步修改意见。
- **若用户明确要求执行**（如回复"执行测试"、"确认运行"、"运行"、"执行"）：进入阶段六。
- **此阻断机制适用于所有场景，无任何例外。严禁在阶段四完成后自动执行测试。**

### 阶段六：测试执行与交付（用户确认执行后）
1. **执行测试**：
   - 根据方案中的执行命令，在终端/命令行中运行集成测试套件
   - 捕获并展示测试结果（通过数、失败数、错误日志）
   - 若测试失败，分析失败原因，输出诊断报告，不擅自修改被测业务代码来"让测试通过"

2. **输出《集成测试执行报告》**：
   - 执行命令
   - 通过率
   - 失败用例及原因分析（如有）
   - 环境启动耗时

3. **输出《规范符合性声明》**：
   - 声明已遵循代码规范文档的关键条款
   - 声明环境隔离策略已落实

4. **询问用户**：
   - 是否需要补充性能测试、压力测试或端到端（E2E）测试
   - 是否需要配置 CI/CD 流水线自动执行集成测试
   - 是否需要根据测试结果调整测试用例或修复被测代码

## 集成测试策略规范

### 1. 命名规范
- 测试类：`XxxIntegrationTest` / `test_xxx_integration.py`
- 测试方法：
  - Java：`shouldXxxWhenXxx` / `testXxxXxx`
  - Python：`test_xxx_xxx_integration` / `test_should_xxx_when_xxx`
  - JS/TS：`should xxx when xxx (integration)`

### 2. 结构规范（AAA + 环境生命周期）
```go
package integration_test

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/mysql"
)

func TestUserRegistration_Integration(t *testing.T) {
    ctx := context.Background()
    
    // Arrange: 启动MySQL容器
    mysqlContainer, err := mysql.Run(ctx, "mysql:8.0",
        mysql.WithDatabase("testdb"),
        mysql.WithUsername("test"),
        mysql.WithPassword("test"),
    )
    if err != nil {
        t.Fatalf("failed to start container: %s", err)
    }
    defer mysqlContainer.Terminate(ctx)
    
    // 获取连接字符串
    connStr, err := mysqlContainer.ConnectionString(ctx)
    if err != nil {
        t.Fatalf("failed to get connection string: %s", err)
    }
    
    // 初始化数据库连接和HTTP处理器
    db := setupDatabase(t, connStr)
    defer db.Close()
    
    handler := NewUserHandler(db)
    router := setupRouter(handler)
    
    // Act: 发送HTTP请求
    body := `{"email":"a@b.com","name":"Alice","password":"pass123"}`
    req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    
    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)
    
    // Assert: HTTP层
    assert.Equal(t, http.StatusCreated, rec.Code)
    
    // Assert: 数据库层
    var user User
    err = db.Where("email = ?", "a@b.com").First(&user).Error
    assert.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)
}
```

### 3. 覆盖维度（必须覆盖）
| 维度 | 说明 | 示例 |
|------|------|------|
| 正常链路 | 标准请求贯穿全链路 | HTTP → Handler → Service → DB → 响应 |
| 异常链路 | 下游服务异常时的行为 | 数据库连接断开、Kafka 不可用、第三方超时 |
| 数据一致性 | 事务与持久化验证 | 注册后数据库有记录、缓存已更新、消息已投递 |
| 并发/时序 | 多请求竞争或顺序依赖 | 重复提交、乐观锁冲突、消息消费顺序 |
| 边界与契约 | 接口契约严格匹配 | 字段缺失、类型不匹配、非法枚举值 |
| 清理验证 | 测试结束后无脏数据 | TRUNCATE 成功、容器状态重置 |

### 4. 环境隔离规范
- **数据库**：必须使用独立 Schema 或 Testcontainers 实例，禁止连接开发/生产库
- **缓存**：使用独立 DB 索引或 Embedded/Container 实例
- **消息队列**：使用独立 Topic/Queue 或 Container 实例
- **第三方 HTTP 服务**：使用 httptest.Server 或 mockhttp 模拟，禁止调用真实付费 API
- **Go 生态集成测试工具推荐**：
  - **testcontainers-go**：官方容器测试库，支持 MySQL/PostgreSQL/Redis/Kafka 等
  - **dockertest**：Docker 容器测试，API 简洁
  - **httptest**：HTTP 服务测试，标准库自带
  - **miniredis**：纯 Go 实现的 Redis 模拟，无需 Docker
  - **ginkgo**：BDD 风格测试框架，适合复杂集成测试
  - **gock**：HTTP 请求 Mock，用于模拟第三方 API

### 5. 数据清理策略
| 策略 | 适用场景 | 实现方式 |
|------|---------|---------|
| 事务回滚 | 单测内联数据库操作 | `tx.Rollback()` 或测试事务包装 |
| 表清理 | 跨事务或非事务操作 | `TRUNCATE TABLE` / `DELETE FROM` |
| 容器重启 | 状态复杂或污染严重 | 每个测试函数重启容器 |
| 数据工厂 | 需要复杂关联数据 | Factory 函数 / Fixture 文件 |

## 禁止事项
- ❌ 禁止在阶段一完成前跳过文档读取，进入任何后续阶段。
- ❌ 禁止在阶段二完成前生成任何测试代码或环境配置。
- ❌ 禁止在阶段三（第一阻断点）用户未确认前执行阶段四。
- ❌ 禁止在阶段四完成后自动执行测试，必须等待阶段五（第二阻断点）用户确认。
- ❌ 禁止在阶段五（第二阻断点）用户未确认执行前，擅自运行测试命令。
- ❌ 禁止集成测试连接生产环境或开发环境的数据库、缓存、消息队列。
- ❌ 禁止在集成测试中验证单个函数的内部逻辑（那是单元测试的职责）。
- ❌ 禁止测试失败后擅自修改被测业务代码来"凑"测试通过。
- ❌ 禁止使用项目规范中未列出的技术栈或命名风格。
- ❌ 禁止遗漏环境清理策略，导致测试间数据污染。
- ❌ 禁止在测试代码中硬编码生产环境的 URL、密码、密钥。

## 引用资源
- 本技能执行时，应将项目内的 PRD、架构概要、API 契约、数据库设计、代码规范文档作为首要上下文参考。
- 若项目文档路径与上述默认路径不一致，优先以用户当前对话中指定的路径为准。
- 优先沿用项目已有的测试框架、容器化工具和 CI/CD 配置。
