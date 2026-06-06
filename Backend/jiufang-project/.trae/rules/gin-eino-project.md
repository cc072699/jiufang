---
alwaysApply: false
description: "Gin + Eino AI Agent 项目。当开发涉及 AI Agent、LLM 集成、多步骤工作流的智能应用时使用此规则。"
---

# Gin + Eino AI Agent 项目规范

本项目是一个 AI Agent 应用，使用 Eino 框架构建智能体工作流。

> **注意**：项目结构在 `gin-base.md` 基础上新增 `agent/` 和 `callback/` 目录。

## 项目类型

- AI Agent 应用
- LLM 集成
- 多步骤工作流
- 工具调用
- 记忆管理

## 新增目录结构

```text
internal/
├── agent/
│   ├── graph.go          # Graph 编排
│   ├── react.go          # ReAct Agent
│   ├── tools.go          # 工具定义
│   └── state.go          # 状态定义
├── callback/
│   └── tracing.go        # Eino Callbacks
└── ... (其他目录同 gin-base.md)
```

## Eino 核心组件

| 组件 | 用途 | 包路径 |
|------|------|--------|
| ChatModel | LLM 交互抽象 | `github.com/cloudwego/eino/components/model` |
| Tool | 工具调用 | `github.com/cloudwego/eino/components/tool` |
| ChatTemplate | 提示模板 | `github.com/cloudwego/eino/components/prompt` |
| Retriever | 知识检索 | `github.com/cloudwego/eino/components/retriever` |
| Embedding | 向量嵌入 | `github.com/cloudwego/eino/components/embedding` |

## LLM 提供商配置

### OpenAI

```go
package model

import (
    "context"
    "os"

    openai "github.com/cloudwego/eino-ext/components/model/openai"
)

func NewOpenAIModel(ctx context.Context) (*openai.ChatModel, error) {
    return openai.NewChatModel(ctx, &openai.ChatModelConfig{
        Model:   "gpt-4o",
        APIKey:  os.Getenv("OPENAI_API_KEY"),
        BaseURL: os.Getenv("OPENAI_BASE_URL"),
    })
}
```

### ARK (ByteDance)

```go
package model

import (
    "context"
    "os"

    ark "github.com/cloudwego/eino-ext/components/model/ark"
)

func NewARKModel(ctx context.Context) (*ark.ChatModel, error) {
    return ark.NewChatModel(ctx, &ark.ChatModelConfig{
        Model:      "doubao-pro-32k",
        APIKey:     os.Getenv("ARK_API_KEY"),
        EndpointID: os.Getenv("ARK_ENDPOINT_ID"),
    })
}
```

## 工具定义

```go
package tool

import (
    "context"

    "github.com/cloudwego/eino/schema"
)

type WeatherInput struct {
    City string `json:"city" jsonschema:"description=城市名称"`
}

type WeatherOutput struct {
    Temperature string `json:"temperature"`
    Condition   string `json:"condition"`
}

func NewWeatherTool() *schema.Tool {
    return &schema.Tool{
        Type: schema.ToolTypeFunction,
        Function: schema.FunctionDefinition{
            Name:        "get_weather",
            Description: "获取指定城市的天气信息",
            Parameters: schema.Parameters{
                Type: schema.DataTypeObject,
                Properties: map[string]*schema.Parameter{
                    "city": {
                        Type:        schema.DataTypeString,
                        Description: "城市名称",
                    },
                },
                Required: []string{"city"},
            },
        },
    }
}

func ExecuteWeatherTool(ctx context.Context, input *WeatherInput) (*WeatherOutput, error) {
    return &WeatherOutput{
        Temperature: "25°C",
        Condition:   "晴天",
    }, nil
}
```

## ReAct Agent

```go
package agent

import (
    "context"

    "github.com/cloudwego/eino/components/model"
    "github.com/cloudwego/eino/components/tool"
    "github.com/cloudwego/eino/flow/agent/react"
    "github.com/cloudwego/eino/schema"
)

type Agent struct {
    agent *react.Agent
}

func NewReActAgent(chatModel model.ChatModel, tools []tool.BaseTool) (*Agent, error) {
    agent, err := react.NewAgent(context.Background(), &react.AgentConfig{
        Model: chatModel,
        Tools: tools,
    })
    if err != nil {
        return nil, err
    }
    return &Agent{agent: agent}, nil
}

func (a *Agent) Run(ctx context.Context, input string) (*schema.Message, error) {
    return a.agent.Invoke(ctx, []*schema.Message{
        schema.UserMessage(input),
    })
}

func (a *Agent) Stream(ctx context.Context, input string) (*schema.StreamReader[*schema.Message], error) {
    return a.agent.Stream(ctx, []*schema.Message{
        schema.UserMessage(input),
    })
}
```

## Graph 编排

```go
package agent

import (
    "context"
    "fmt"

    "github.com/cloudwego/eino/components/model"
    "github.com/cloudwego/eino/components/tool"
    "github.com/cloudwego/eino/compose"
    "github.com/cloudwego/eino/schema"
)

type GraphState struct {
    Input   string
    Output  string
    History []*schema.Message
}

func NewAgentGraph(chatModel model.ChatModel, tools []tool.BaseTool) (*compose.Graph[*GraphState, *GraphState], error) {
    g := compose.NewGraph[*GraphState, *GraphState]()

    g.AddNode("chat", func(ctx context.Context, state *GraphState) (*GraphState, error) {
        msg, err := chatModel.Generate(ctx, []*schema.Message{
            schema.UserMessage(state.Input),
        })
        if err != nil {
            return nil, fmt.Errorf("chat failed: %w", err)
        }
        state.History = append(state.History, msg)
        return state, nil
    })

    g.AddNode("tools", func(ctx context.Context, state *GraphState) (*GraphState, error) {
        return state, nil
    })

    g.AddEdge(compose.START, "chat")
    g.AddEdge("chat", "tools")
    g.AddEdge("tools", compose.END)

    return g, nil
}
```

## Callbacks 追踪

```go
package callback

import (
    "context"

    "github.com/cloudwego/eino/callbacks"
    "go.uber.org/zap"
)

type TracingCallback struct {
    callbacks.Handler
    logger *zap.Logger
}

func NewTracingCallback(logger *zap.Logger) *TracingCallback {
    return &TracingCallback{logger: logger}
}

func (t *TracingCallback) OnStart(ctx context.Context, info *callbacks.RunInfo, input any) context.Context {
    t.logger.Info("run started",
        zap.String("run_id", info.RunID),
        zap.String("name", info.Name),
    )
    return ctx
}

func (t *TracingCallback) OnEnd(ctx context.Context, info *callbacks.RunInfo, output any) context.Context {
    t.logger.Info("run ended",
        zap.String("run_id", info.RunID),
        zap.String("name", info.Name),
    )
    return ctx
}

func (t *TracingCallback) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
    t.logger.Error("run error",
        zap.String("run_id", info.RunID),
        zap.Error(err),
    )
    return ctx
}

func (t *TracingCallback) OnStream(ctx context.Context, info *callbacks.RunInfo, chunk any) context.Context {
    t.logger.Debug("stream chunk",
        zap.String("run_id", info.RunID),
    )
    return ctx
}
```

## ChatTemplate

```go
package prompt

import (
    "github.com/cloudwego/eino/components/prompt"
    "github.com/cloudwego/eino/schema"
)

func NewSystemPrompt() *prompt.ChatTemplate {
    return prompt.FromMessages(schema.SystemMessage(
        "你是一个专业的AI助手，请根据用户的问题提供准确、有帮助的回答。",
    ))
}

func NewChatPrompt(systemPrompt, userMessage string) *prompt.ChatTemplate {
    return prompt.FromMessages(
        schema.SystemMessage(systemPrompt),
        schema.UserMessage(userMessage),
    )
}
```

## 记忆存储 (pgvector)

```go
package memory

import (
    "context"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/pgvector/pgvector-go"
)

type MemoryStore struct {
    db *pgxpool.Pool
}

func NewMemoryStore(connString string) (*MemoryStore, error) {
    ctx := context.Background()
    pool, err := pgxpool.New(ctx, connString)
    if err != nil {
        return nil, err
    }
    return &MemoryStore{db: pool}, nil
}

func (m *MemoryStore) Add(ctx context.Context, userID string, content string, embedding []float32) error {
    query := `INSERT INTO memories (user_id, content, embedding) VALUES ($1, $2, $3)`
    _, err := m.db.Exec(ctx, query, userID, content, pgvector.NewVector(embedding))
    return err
}

func (m *MemoryStore) Search(ctx context.Context, userID string, embedding []float32, limit int) ([]string, error) {
    query := `
        SELECT content FROM memories
        WHERE user_id = $1
        ORDER BY embedding <=> $2
        LIMIT $3
    `
    rows, err := m.db.Query(ctx, query, userID, pgvector.NewVector(embedding), limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var results []string
    for rows.Next() {
        var content string
        if err := rows.Scan(&content); err != nil {
            return nil, err
        }
        results = append(results, content)
    }
    return results, nil
}

func (m *MemoryStore) Close() {
    m.db.Close()
}
```

## 依赖列表

| 包名 | 用途 | 导入路径 |
|------|------|----------|
| Gin | Web框架 | `github.com/gin-gonic/gin` |
| Eino | Agent框架 | `github.com/cloudwego/eino` |
| Eino OpenAI | OpenAI模型 | `github.com/cloudwego/eino-ext/components/model/openai` |
| Eino ARK | 字节模型 | `github.com/cloudwego/eino-ext/components/model/ark` |
| GORM | ORM | `gorm.io/gorm` |
| pgx | PostgreSQL驱动 | `github.com/jackc/pgx/v5` |
| pgvector | 向量扩展 | `github.com/pgvector/pgvector-go` |
| Viper | 配置管理 | `github.com/spf13/viper` |
| JWT | 认证 | `github.com/golang-jwt/jwt/v5` |

## 关键约定

1. 所有 LLM 操作必须有追踪 (Callbacks)
2. 所有 Agent 必须支持流式响应
3. 所有工具必须有清晰的参数定义
4. 所有记忆按 user_id 隔离
5. 所有 LLM 调用使用 context 控制超时
