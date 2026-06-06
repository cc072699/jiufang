package user

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
	SnowflakeID  int64  `gorm:"uniqueIndex;not null" json:"snowflake_id,string"`
	Username     string `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password     string `gorm:"size:255;not null" json:"-"`
	Email        string `gorm:"uniqueIndex;size:100;not null" json:"email"`
	Avatar       string `gorm:"size:255" json:"avatar"`
	Role         string `gorm:"size:20;not null" json:"role"`
	Status       int    `gorm:"default:1" json:"status"`
	IsFirstLogin bool   `gorm:"default:true" json:"is_first_login"`
}

func (User) TableName() string {
	return "users"
}

type Role string

const (
	RoleAdmin     Role = "admin"
	RoleManager   Role = "manager"
	RoleExecutive Role = "executive"
)

type UserStatus int

const (
	StatusDisabled UserStatus = 0
	StatusEnabled  UserStatus = 1
)

func IsValidRole(role string) bool {
	return role == string(RoleAdmin) || role == string(RoleManager) || role == string(RoleExecutive)
}

func IsValidStatus(status int) bool {
	return status == int(StatusDisabled) || status == int(StatusEnabled)
}
