# Go Gin + Eino AI 编程规范体系

一套完整的AI编程规范体系，用于指导AI助手（如Trae）进行Go Gin + Eino项目的开发工作。

## 项目概述

本项目定义了一套完整的AI编程规范体系，包括：

- **规则文件（Rules）**：定义Go语言、Gin框架、Eino框架的编码规范
- **技能文件（Skills）**：定义从需求到测试的完整开发工作流
- **配置文件**：定义开发环境、代码检查、CI/CD等工具配置

## 目录结构

```
.
├── .trae/
│   ├── rules/                          # 规则文件目录
│   │   ├── go-base.md                  # Go语言基础规范
│   │   ├── gin-base.md                 # Gin框架基础规范
│   │   ├── gin-project.md              # 纯Gin项目规范
│   │   └── gin-eino-project.md         # Gin+Eino项目规范
│   └── skills/                         # 技能文件目录
│       ├── prd-generator/              # PRD文档生成
│       ├── architecture-overview-generator/  # 架构概要生成
│       ├── detailed-design-generator/  # 详细设计生成
│       ├── module-developer/           # 功能开发
│       ├── database-migrator/          # 数据库迁移
│       ├── unit-test-generator/        # 单元测试生成
│       ├── integration-test-developer/ # 集成测试开发
│       ├── api-doc-generator/          # API文档生成
│       ├── code-reviewer/              # 代码审查
│       ├── bug-fixer/                  # Bug修复
│       └── test-case-generator/        # 测试用例生成
├── .golangci.yml                       # golangci-lint配置
├── .pre-commit-config.yaml             # pre-commit配置
├── .gitignore                          # Git忽略配置
├── .secrets.baseline                   # detect-secrets基线
├── Jenkinsfile                         # Jenkins CI/CD配置
├── Makefile                            # Make构建配置
├── SETUP.md                            # 开发环境安装指南
└── README.md                           # 本文件
```

## 规则文件说明

### go-base.md

Go语言基础规范，包括：
- 命名规范（包名、变量、常量、接口）
- 目录结构规范
- 错误处理规范
- Context使用规范
- 并发规范
- 日志规范（使用zap）
- 测试规范
- 代码风格规范

### gin-base.md

Gin框架基础规范，包括：
- 项目结构规范
- 统一响应格式
- Handler模式
- 请求验证
- 中间件规范（Auth、Logger、Recovery、CORS）
- 路由组织
- 配置管理
- 日志初始化（zap）
- 监控规范（Prometheus）

### gin-project.md

纯Gin Web API项目规范，包括：
- GORM模型定义
- Repository模式
- 事务处理
- Service层
- 错误处理
- 重试逻辑

### gin-eino-project.md

Gin + Eino AI Agent项目规范，包括：
- Eino核心组件（ChatModel、Tool、Agent、Graph、Memory）
- LLM提供商配置（OpenAI、ARK）
- 工具定义
- ReAct Agent创建
- Graph编排
- Callbacks追踪
- ChatTemplate
- 记忆存储（pgvector）

## 技能文件说明

| 技能 | 触发场景 | 功能 |
|------|----------|------|
| prd-generator | 生成PRD文档 | 生成产品需求文档（八章结构） |
| architecture-overview-generator | 生成架构概要 | 生成技术架构概要文档（六章结构） |
| detailed-design-generator | 生成详细设计 | 生成技术详细设计文档（八章结构） |
| module-developer | 功能开发/修改 | 执行功能开发与修改任务 |
| database-migrator | 数据库迁移 | 生成数据库迁移脚本和回滚脚本 |
| unit-test-generator | 单元测试 | 为现有代码生成单元测试 |
| integration-test-developer | 集成测试 | 搭建集成测试环境并编写测试 |
| api-doc-generator | API文档生成 | 生成OpenAPI/Swagger文档 |
| code-reviewer | 代码审查 | 进行代码质量和规范检查 |
| bug-fixer | Bug修复 | 分析并修复Bug |
| test-case-generator | 测试用例生成 | 为测试人员生成测试用例 |

## 开发工作流

```
PRD文档 → 架构概要 → 详细设计 → 功能开发 → 单元测试/集成测试 → API文档 → 代码审查
    ↓
测试用例生成（测试人员使用）
```

### 文档生成流程

1. **需求阶段**：使用 `prd-generator` 生成PRD文档
2. **设计阶段**：使用 `architecture-overview-generator` 生成架构概要
3. **详细设计**：使用 `detailed-design-generator` 生成详细设计文档
4. **开发阶段**：使用 `module-developer` 进行功能开发
5. **测试阶段**：使用 `unit-test-generator` 和 `integration-test-developer` 编写测试
6. **文档阶段**：使用 `api-doc-generator` 生成API文档
7. **审查阶段**：使用 `code-reviewer` 进行代码审查

### 异常流程

- **Bug修复**：使用 `bug-fixer` 进行问题定位和修复
- **数据库变更**：使用 `database-migrator` 生成迁移脚本

## 技术栈

| 类别 | 技术选型 |
|------|----------|
| 语言 | Go 1.21+ |
| Web框架 | Gin |
| AI框架 | Eino（字节跳动） |
| ORM | GORM |
| 日志 | Zap |
| 配置 | Viper |
| 监控 | Prometheus |
| 测试 | Go testing + testify + gomock + testcontainers-go |
| 数据库迁移 | golang-migrate |
| API文档 | OpenAPI 3.0 + swaggo |
| CI/CD | Jenkins |
| 代码检查 | golangci-lint |
| Git钩子 | pre-commit |

## 快速开始

### 1. 安装前置条件

确保已安装以下工具：
- Go 1.21+
- Git
- make
- uv（Python包管理器）

### 2. 配置开发环境

```bash
# 检查前置条件
make check-prerequisites

# 安装开发工具
make setup
```

详细安装说明请参考 [SETUP.md](SETUP.md)

### 3. 常用命令

```bash
make build        # 构建项目
make test         # 运行测试
make lint         # 代码检查
make check        # 运行所有检查
make help         # 显示所有命令
```

## 使用规范

### 在Trae IDE中使用

1. **规则自动加载**：将此项目克隆到你的Go项目根目录，Trae会自动加载`.trae/rules/`下的规则文件

2. **技能按需触发**：当你向AI提出相关需求时，Trae会根据技能的`description`字段自动判断并加载对应的技能

3. **手动调用技能**：你也可以明确告诉AI使用某个技能，例如：
   - "使用prd-generator生成PRD文档"
   - "使用code-reviewer审查这段代码"

### 自定义规范

1. **修改规则文件**：编辑`.trae/rules/`下的文件，添加项目特定的规范

2. **修改技能文件**：编辑`.trae/skills/{skill-name}/SKILL.md`，调整工作流程或检查项

3. **添加新技能**：在`.trae/skills/`下创建新的技能目录和SKILL.md文件

## 角色使用指南

本项目为不同角色提供了详细的使用说明文档，请根据您的角色查阅对应的指南：

| 角色 | 使用指南 | 核心技能 |
|------|----------|----------|
| 产品经理 | [docs/user-guide-product-manager.md](docs/user-guide-product-manager.md) | prd-generator |
| 架构师 | [docs/user-guide-architect.md](docs/user-guide-architect.md) | architecture-overview-generator, detailed-design-generator |
| 开发人员 | [docs/user-guide-developer.md](docs/user-guide-developer.md) | module-developer, database-migrator, unit-test-generator, integration-test-developer, bug-fixer, code-reviewer |
| 测试人员 | [docs/user-guide-tester.md](docs/user-guide-tester.md) | test-case-generator |
| 运维人员 | [docs/user-guide-ops.md](docs/user-guide-ops.md) | Jenkins管理、数据库迁移执行、监控配置 |

**各角色职责边界**：

- **产品经理**：定义需求，生成PRD文档
- **架构师**：技术设计，生成架构概要和详细设计
- **开发人员**：功能实现，编写测试代码，代码审查，Bug修复
- **测试人员**：设计测试用例，执行接口测试，报告Bug（不接触代码）
- **运维人员**：服务器运维，Jenkins构建管理，执行数据库迁移（不编写代码）

## 最佳实践

1. **文档先行**：在开发前先生成PRD、架构概要、详细设计文档
2. **测试驱动**：使用unit-test-generator和integration-test-developer保证代码质量
3. **代码审查**：在提交代码前使用code-reviewer进行审查
4. **规范遵循**：遵循go-base.md和gin-base.md中的编码规范
5. **日志规范**：统一使用zap框架，遵循日志级别和埋点规范

## 贡献指南

欢迎贡献新的规范、技能或改进建议：

1. Fork本项目
2. 创建特性分支
3. 提交变更
4. 创建Pull Request

## 许可证

MIT License
