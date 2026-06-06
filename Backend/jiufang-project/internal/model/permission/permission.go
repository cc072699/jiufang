package permission

import (
	"time"

	"gorm.io/gorm"
)

// Permission represents a data access permission for a user group.
type Permission struct {
	ID              uint           `gorm:"primaryKey" json:"-"`
	SnowflakeID     int64          `gorm:"uniqueIndex;not null" json:"id,string"`
	GroupID         uint           `gorm:"not null;index" json:"group_id"`
	TableName       string         `gorm:"column:table_name;size:100;not null" json:"table_name"` // 表名
	AllowedFields   string         `gorm:"type:text" json:"allowed_fields"`                       // 允许查询的字段列表（JSON数组格式）
	FilterCondition string         `gorm:"size:500" json:"filter_condition"`                      // 数据级权限过滤条件（SQL条件表达式）
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// Note: TableName field is a business field, not the GORM table name method.
// GORM will use the default table name "permissions" (plural of struct name).
