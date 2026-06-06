---
alwaysApply: false
description: "纯 Gin Web API 项目。当开发不涉及 AI Agent 的后端 Web API 服务时使用此规则。"
---

# 纯 Gin Web API 项目规范

本项目是一个标准的 Gin Web API 服务，不涉及 AI Agent 功能。

> **注意**：项目结构遵循 `gin-base.md` 中定义的标准结构，此处不再重复。

## 项目类型

- 纯后端 REST API 服务
- 无 AI Agent 功能
- 无 LLM 集成

## 数据库操作

### GORM 模型

```go
package model

import (
    "time"

    "gorm.io/gorm"
)

type Model struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type User struct {
    Model
    Username string `gorm:"uniqueIndex;size:50;not null" json:"username"`
    Email    string `gorm:"uniqueIndex;size:100;not null" json:"email"`
    Password string `gorm:"size:255;not null" json:"-"`
    Status   int    `gorm:"default:1" json:"status"`
}
```

### Repository 模式

```go
package repository

import (
    "context"

    "gorm.io/gorm"
)

type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
    var user model.User
    err := r.db.WithContext(ctx).First(&user, id).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
    var user model.User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]model.User, int64, error) {
    var users []model.User
    var total int64

    if err := r.db.WithContext(ctx).Model(&model.User{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }

    err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&users).Error
    return users, total, err
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
    return r.db.WithContext(ctx).Save(user).Error
}

func (r *UserRepository) Delete(ctx context.Context, id uint) error {
    return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}
```

### 事务处理

```go
func (s *Service) Transfer(ctx context.Context, fromID, toID uint, amount float64) error {
    return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        if err := tx.Model(&model.Account{}).Where("id = ?", fromID).
            Update("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
            return err
        }

        if err := tx.Model(&model.Account{}).Where("id = ?", toID).
            Update("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
            return err
        }

        return nil
    })
}
```

## Service 层

```go
package service

import (
    "context"
    "fmt"

    "golang.org/x/crypto/bcrypt"
)

type UserService struct {
    repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
    return &UserService{repo: repo}
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

func (s *UserService) Create(ctx context.Context, req *CreateUserRequest) (*UserResponse, error) {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, fmt.Errorf("failed to hash password: %w", err)
    }

    user := &model.User{
        Username: req.Username,
        Email:    req.Email,
        Password: string(hashedPassword),
    }

    if err := s.repo.Create(ctx, user); err != nil {
        return nil, fmt.Errorf("create user failed: %w", err)
    }

    return &UserResponse{
        ID:       user.ID,
        Username: user.Username,
        Email:    user.Email,
    }, nil
}

func (s *UserService) GetByID(ctx context.Context, id uint) (*UserResponse, error) {
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get user failed: %w", err)
    }

    return &UserResponse{
        ID:       user.ID,
        Username: user.Username,
        Email:    user.Email,
    }, nil
}
```

## 错误处理

```go
package errors

import "errors"

var (
    ErrInvalidRequest     = errors.New("invalid request")
    ErrUnauthorized       = errors.New("unauthorized")
    ErrNotFound           = errors.New("not found")
    ErrInternalError      = errors.New("internal error")
)

type AppError struct {
    Code    int
    Message string
    Err     error
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return e.Message + ": " + e.Err.Error()
    }
    return e.Message
}

func (e *AppError) Unwrap() error {
    return e.Err
}

func NewAppError(code int, message string, err error) *AppError {
    return &AppError{Code: code, Message: message, Err: err}
}

func IsAppError(err error) bool {
    var appErr *AppError
    return errors.As(err, &appErr)
}
```

## 重试逻辑

```go
package retry

import (
    "context"
    "time"

    "github.com/avast/retry-go/v4"
)

func Do(ctx context.Context, operation func() error, opts ...retry.Option) error {
    defaultOpts := []retry.Option{
        retry.Attempts(3),
        retry.Delay(time.Second * 2),
        retry.DelayType(retry.BackOffDelay),
        retry.Context(ctx),
    }
    return retry.Do(operation, append(defaultOpts, opts...)...)
}
```

## 依赖列表

| 包名 | 用途 | 导入路径 |
|------|------|----------|
| Gin | Web框架 | `github.com/gin-gonic/gin` |
| GORM | ORM | `gorm.io/gorm` |
| PostgreSQL Driver | 数据库驱动 | `gorm.io/driver/postgres` |
| Viper | 配置管理 | `github.com/spf13/viper` |
| JWT | 认证 | `github.com/golang-jwt/jwt/v5` |
| Validator | 请求验证 | `github.com/go-playground/validator/v10` |
| Retry | 重试逻辑 | `github.com/avast/retry-go/v4` |
| Tollbooth | 限流 | `github.com/didip/tollbooth` |
| Testify | 测试 | `github.com/stretchr/testify` |
| Bcrypt | 密码加密 | `golang.org/x/crypto/bcrypt` |

## 关键约定

1. 所有数据库操作必须使用 context
2. 所有 Service 方法返回错误
3. 所有 Repository 使用 GORM
4. 所有事务在 Service 层处理
5. 所有密码必须使用 bcrypt 加密存储
