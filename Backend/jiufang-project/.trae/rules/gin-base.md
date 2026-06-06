---
alwaysApply: false
description: 
---
# Gin 框架基础规范

所有使用 Gin 框架的项目必须遵循以下规范。

## 项目结构

```text
cmd/
└── server/
    └── main.go              # 应用入口
internal/
├── api/
│   └── v1/
│       ├── handler.go       # HTTP处理器
│       └── routes.go        # 路由定义
├── config/
│   └── config.go            # 配置加载
├── middleware/
│   ├── auth.go              # 认证中间件
│   ├── logger.go            # 日志中间件
│   └── recovery.go          # 恢复中间件
├── model/
│   └── models.go            # 数据模型
├── repository/
│   └── repository.go        # 数据访问层
├── service/
│   └── service.go           # 业务逻辑层
└── pkg/
    ├── logger/
    ├── response/
    └── errors/
configs/
├── config.yaml
└── config.example.yaml
go.mod
go.sum
```

## 统一响应格式

```go
package response

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

type PageResponse struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    Total   int64       `json:"total"`
    Page    int         `json:"page"`
    Size    int         `json:"size"`
}

func Success(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, Response{
        Code:    0,
        Message: "success",
        Data:    data,
    })
}

func Error(c *gin.Context, code int, message string) {
    c.JSON(http.StatusOK, Response{
        Code:    code,
        Message: message,
    })
}

func Page(c *gin.Context, data interface{}, total int64, page, size int) {
    c.JSON(http.StatusOK, PageResponse{
        Code:    0,
        Message: "success",
        Data:    data,
        Total:   total,
        Page:    page,
        Size:    size,
    })
}
```

## Handler 模式

```go
package handler

import (
    "github.com/gin-gonic/gin"
    "github.com/go-playground/validator/v10"
)

type UserHandler struct {
    service   UserService
    validator *validator.Validate
}

func NewUserHandler(service UserService) *UserHandler {
    return &UserHandler{
        service:   service,
        validator: validator.New(),
    }
}

type CreateUserRequest struct {
    Username string `json:"username" validate:"required,min=3,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8,max=100"`
}

type UserResponse struct {
    ID       uint   `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
}

func (h *UserHandler) Create(c *gin.Context) {
    ctx := c.Request.Context()

    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, 400, "invalid request body")
        return
    }

    if err := h.validator.Struct(req); err != nil {
        response.Error(c, 400, err.Error())
        return
    }

    result, err := h.service.Create(ctx, &req)
    if err != nil {
        response.Error(c, 500, err.Error())
        return
    }

    response.Success(c, result)
}
```

## 请求验证

```go
type PageRequest struct {
    Page int `form:"page" binding:"min=1"`
    Size int `form:"size" binding:"min=1,max=100"`
}

func (p *PageRequest) Offset() int {
    if p.Page <= 0 {
        return 0
    }
    return (p.Page - 1) * p.Size
}

func (p *PageRequest) Default() *PageRequest {
    if p.Page <= 0 {
        p.Page = 1
    }
    if p.Size <= 0 {
        p.Size = 10
    }
    return p
}
```

## 中间件规范

### 认证中间件

```go
package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

func Auth(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
            c.Abort()
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        if tokenString == authHeader {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token format"})
            c.Abort()
            return
        }

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, jwt.ErrSignatureInvalid
            }
            return []byte(jwtSecret), nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            c.Abort()
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
            c.Abort()
            return
        }

        c.Set("user_id", claims["sub"])
        c.Next()
    }
}

func GetUserID(c *gin.Context) string {
    if userID, exists := c.Get("user_id"); exists {
        if id, ok := userID.(string); ok {
            return id
        }
    }
    return ""
}
```

### 日志中间件

```go
package middleware

import (
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path

        c.Next()

        latency := time.Since(start)
        logger.Log.Info("request",
            zap.String("method", c.Request.Method),
            zap.String("path", path),
            zap.Int("status", c.Writer.Status()),
            zap.Duration("latency", latency),
        )
    }
}
```

### 恢复中间件

```go
package middleware

import (
    "net/http"
    "runtime/debug"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

func Recovery() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                logger.Log.Error("panic recovered",
                    zap.Any("error", err),
                    zap.String("stack", string(debug.Stack())),
                )
                c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
                c.Abort()
            }
        }()
        c.Next()
    }
}
```

### CORS 中间件

```go
package middleware

import (
    "github.com/gin-gonic/gin"
)

func CORS(allowOrigins []string) gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")
        for _, o := range allowOrigins {
            if o == "*" || o == origin {
                c.Header("Access-Control-Allow-Origin", o)
                break
            }
        }
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    }
}
```

## 路由组织

```go
package api

import "github.com/gin-gonic/gin"

type Handlers struct {
    User  *handler.UserHandler
    // 其他 Handler...
}

func SetupRoutes(r *gin.Engine, handlers *Handlers) {
    api := r.Group("/api/v1")
    {
        users := api.Group("/users")
        {
            users.POST("", handlers.User.Create)
            users.GET("/:id", handlers.User.GetByID)
            users.GET("", handlers.User.List)
            users.PUT("/:id", handlers.User.Update)
            users.DELETE("/:id", handlers.User.Delete)
        }
    }

    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
}
```

## 配置管理

```go
package config

import "github.com/spf13/viper"

type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Auth     AuthConfig     `mapstructure:"auth"`
}

type ServerConfig struct {
    Port         int    `mapstructure:"port"`
    ReadTimeout  int    `mapstructure:"read_timeout"`
    WriteTimeout int    `mapstructure:"write_timeout"`
    LogLevel     string `mapstructure:"log_level"`
}

type DatabaseConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    User     string `mapstructure:"user"`
    Password string `mapstructure:"password"`
    Database string `mapstructure:"database"`
}

type AuthConfig struct {
    JWTSecret string `mapstructure:"jwt_secret"`
}

func Load(path string) (*Config, error) {
    v := viper.New()
    v.SetConfigFile(path)
    v.AutomaticEnv()

    if err := v.ReadInConfig(); err != nil {
        return nil, err
    }

    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}
```

## 日志初始化

```go
package logger

import (
    "os"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func Init(level string) {
    var config zap.Config
    config = zap.Config{
        Level:       getZapLevel(level),
        Development: false,
        Encoding:    "json",
        EncoderConfig: zapcore.EncoderConfig{
            TimeKey:        "time",
            LevelKey:       "level",
            NameKey:        "logger",
            CallerKey:      "caller",
            MessageKey:     "msg",
            StacktraceKey:  "stacktrace",
            LineEnding:     zapcore.DefaultLineEnding,
            EncodeLevel:    zapcore.LowercaseLevelEncoder,
            EncodeTime:     zapcore.ISO8601TimeEncoder,
            EncodeDuration: zapcore.SecondsDurationEncoder,
            EncodeCaller:   zapcore.ShortCallerEncoder,
        },
        OutputPaths:      []string{"stdout"},
        ErrorOutputPaths: []string{"stderr"},
    }

    Log, _ = config.Build()
}

func getZapLevel(level string) zap.AtomicLevel {
    switch level {
    case "debug":
        return zap.NewAtomicLevelAt(zapcore.DebugLevel)
    case "info":
        return zap.NewAtomicLevelAt(zapcore.InfoLevel)
    case "warn":
        return zap.NewAtomicLevelAt(zapcore.WarnLevel)
    case "error":
        return zap.NewAtomicLevelAt(zapcore.ErrorLevel)
    default:
        return zap.NewAtomicLevelAt(zapcore.InfoLevel)
    }
}
```

## 主入口

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func main() {
    cfg, err := config.Load("configs/config.yaml")
    if err != nil {
        panic(fmt.Sprintf("failed to load config: %v", err))
    }

    logger.Init(cfg.Server.LogLevel)

    dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
        cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
        cfg.Database.Password, cfg.Database.Database)

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic(fmt.Sprintf("failed to connect database: %v", err))
    }

    userRepo := repository.NewUserRepository(db)
    userService := service.NewUserService(userRepo)
    userHandler := handler.NewUserHandler(userService)

    handlers := &api.Handlers{
        User: userHandler,
    }

    router := gin.New()
    router.Use(middleware.Logger())
    router.Use(middleware.Recovery())
    router.Use(middleware.CORS([]string{"*"}))

    api.SetupRoutes(router, handlers)

    srv := &http.Server{
        Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
        Handler:      router,
        ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
        WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
    }

    go func() {
        logger.Log.Info("server started", zap.Int("port", cfg.Server.Port))
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            panic(fmt.Sprintf("server error: %v", err))
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    logger.Log.Info("shutting down server...")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        logger.Log.Error("server shutdown failed", zap.Error(err))
    }

    logger.Log.Info("server exited")
}
```

## 关键约定

1. 所有 Handler 必须验证请求参数
2. 所有响应必须使用统一响应格式
3. 所有错误必须正确处理和返回
4. 所有路由必须使用路由组组织
5. 所有中间件必须正确处理异常
6. JWT Token 必须使用 `Bearer ` 前缀格式

## 日志规范

### 日志框架选择

| 框架 | 特点 | 适用场景 |
|------|------|----------|
| **zap** | 高性能、结构化、级别丰富 | 生产环境首选 |
| **logrus** | API友好、Hook丰富 | 中小项目 |
| **slog** | 标准库、零依赖 | Go 1.21+项目 |

### 日志级别使用规范

| 级别 | 使用场景 | 示例 |
|------|----------|------|
| DEBUG | 开发调试信息 | 变量值、流程跟踪 |
| INFO | 正常业务事件 | 请求处理、定时任务执行 |
| WARN | 潜在问题 | 慢查询、重试、降级 |
| ERROR | 错误但可恢复 | 业务异常、外部服务失败 |
| FATAL | 致命错误 | 启动失败、配置缺失 |

### 日志埋点规范

- 每个请求必须有唯一 `trace_id`，通过中间件注入
- 关键操作必须记录日志：用户登录、订单创建、支付完成等
- 错误日志必须包含错误详情，ERROR级别需包含堆栈
- 敏感信息（密码、token）禁止记录

```go
logger.Log.Info("create user request",
    zap.String("trace_id", traceID),
    zap.String("action", "create_user"),
)

logger.Log.Error("create user failed",
    zap.String("trace_id", traceID),
    zap.Error(err),
)
```

## 监控规范

### Prometheus指标定义

| 指标名 | 类型 | 标签 | 说明 |
|--------|------|------|------|
| `http_requests_total` | Counter | method, path, status | HTTP请求总数 |
| `http_request_duration_seconds` | Histogram | method, path | HTTP请求延迟 |
| `http_requests_in_flight` | Gauge | method | 当前处理中的请求数 |
| `db_query_duration_seconds` | Histogram | query_type | 数据库查询延迟 |
| `business_operations_total` | Counter | operation, status | 业务操作计数 |

### 指标暴露

```go
r.GET("/metrics", gin.WrapH(promhttp.Handler()))
r.GET("/health", func(c *gin.Context) {
    c.JSON(200, gin.H{"status": "ok"})
})
```

## 可观测性检查清单

| 检查项 | 要求 |
|--------|------|
| 日志格式 | JSON结构化日志 |
| 日志级别 | DEBUG/INFO/WARN/ERROR/FATAL |
| 请求追踪 | 每个请求有唯一trace_id |
| 关键操作日志 | 用户登录、订单创建、支付完成等 |
| 错误日志 | 包含错误详情和堆栈（ERROR级别） |
| HTTP指标 | QPS、延迟、错误率 |
| 业务指标 | 核心业务操作计数 |
| 健康检查 | /health端点 |
| 指标暴露 | /metrics端点（Prometheus格式） |
