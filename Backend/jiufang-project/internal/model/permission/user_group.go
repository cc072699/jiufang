package permission

import (
	"time"

	"gorm.io/gorm"
)

type UserGroup struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	SnowflakeID int64          `gorm:"uniqueIndex;not null" json:"snowflake_id,string"`
	Name        string         `gorm:"uniqueIndex;size:50;not null" json:"name"`
	Description string         `gorm:"size:200" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (UserGroup) TableName() string {
	return "user_groups"
}

type ResourceType string

const (
	ResourceTypeTable     ResourceType = "table"
	ResourceTypeField     ResourceType = "field"
	ResourceTypeOperation ResourceType = "operation"
)

type PermissionAction string

const (
	ActionRead   PermissionAction = "read"
	ActionWrite  PermissionAction = "write"
	ActionExport PermissionAction = "export"
)

func IsValidResourceType(resourceType string) bool {
	return resourceType == string(ResourceTypeTable) ||
		resourceType == string(ResourceTypeField) ||
		resourceType == string(ResourceTypeOperation)
}

func IsValidPermissionAction(action string) bool {
	return action == string(ActionRead) ||
		action == string(ActionWrite) ||
		action == string(ActionExport)
}
