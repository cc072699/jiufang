---
alwaysApply: false
description: 
---
# Go 语言基础规范

所有 Go 项目必须遵循以下基础规范。

## 命名规范

- **包名**：小写单词，不使用下划线或驼峰（如 `user`, `httpserver`）
- **导出标识符**：PascalCase（如 `UserService`, `GetByID`）
- **私有标识符**：camelCase（如 `userService`, `getByID`）
- **常量**：PascalCase 或全大写下划线分隔（如 `MaxRetry`, `MAX_RETRY`）
- **接口**：动词+er 后缀（如 `Reader`, `Writer`, `UserRepository`）

## 目录结构

- 使用小写字母和下划线（如 `user_handler.go`）
- 测试文件以 `_test.go` 结尾
- 按功能模块组织，不按技术分层

## 错误处理

- **永远不要忽略错误**
- 使用 `if err != nil` 模式
- 使用 `fmt.Errorf` 包装错误上下文
- 自定义错误类型实现 `error` 接口

```go
func (s *Service) DoSomething(ctx context.Context, id string) (*Result, error) {
    if id == "" {
        return nil, errors.New("id is required")
    }

    result, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to find by id: %w", err)
    }

    return result, nil
}
```

## Context 使用

- `context.Context` 作为函数第一个参数
- 不要在结构体中存储 context
- 使用 context 进行超时和取消控制

```go
func (s *Service) Process(ctx context.Context, req *Request) (*Result, error) {
    // 正确：ctx 作为第一个参数
}

// 错误示例
func (s *Service) Process(req *Request) (*Result, error) {
    ctx := context.Background() // 不要这样做
}
```

## 并发规范

- 使用 `go` 关键字启动 goroutine
- 确保 goroutine 能正确退出
- 使用 `sync.WaitGroup` 等待 goroutine 完成
- 使用 `chan` 进行 goroutine 间通信

```go
func (s *Service) ProcessConcurrent(ctx context.Context, items []Item) error {
    var wg sync.WaitGroup
    var mu sync.Mutex
    var firstErr error

    for _, item := range items {
        wg.Add(1)
        go func(i Item) {
            defer wg.Done()
            if err := s.processItem(ctx, i); err != nil {
                mu.Lock()
                if firstErr == nil {
                    firstErr = err
                }
                mu.Unlock()
            }
        }(item)
    }

    wg.Wait()
    return firstErr
}
```

## 日志规范

- 使用结构化日志（**zap** 为生产环境首选）
- 日志消息使用小写
- 包含上下文信息（request_id, user_id 等）

```go
logger.Info("user_created",
    zap.Uint("user_id", user.ID),
    zap.String("username", user.Username),
)

logger.Error("database_error",
    zap.Error(err),
    zap.String("query", query),
)
```

## 测试规范

- 使用 `testing` 包
- 使用 `testify` 进行断言
- 使用表格驱动测试

```go
func TestService_DoSomething(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "success",
            input:   "test",
            want:    "result",
            wantErr: false,
        },
        {
            name:    "empty input",
            input:   "",
            want:    "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

## 代码风格

- 使用 `gofmt` 格式化代码
- 使用 `goimports` 管理导入
- 导入按标准库、第三方库、本地库分组

```go
import (
    "context"
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "myproject/internal/service"
)
```

## 关键原则

1. 简洁优于复杂
2. 显式优于隐式
3. 组合优于继承
4. 接口优于具体类型
5. 错误必须处理
