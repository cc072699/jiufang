// Package audit implements the operation log model.
package audit

import (
	"time"
)

// OperationResult represents the result of an operation.
type OperationResult string

const (
	OperationResultSuccess OperationResult = "success"
	OperationResultFailed  OperationResult = "failed"
)

// OperationType represents the type of an operation.
type OperationType string

const (
	OperationTypeLogin            OperationType = "login"
	OperationTypeLogout           OperationType = "logout"
	OperationTypeQuery            OperationType = "query"
	OperationTypeCreateUser       OperationType = "create_user"
	OperationTypeUpdateUser       OperationType = "update_user"
	OperationTypeDeleteUser       OperationType = "delete_user"
	OperationTypeConfigPermission OperationType = "config_permission"
	OperationTypeCreateReport     OperationType = "create_report"
	OperationTypeUpdateReport     OperationType = "update_report"
	OperationTypeDeleteReport     OperationType = "delete_report"
	OperationTypeCreateAlert      OperationType = "create_alert"
	OperationTypeUpdateAlert      OperationType = "update_alert"
	OperationTypeDeleteAlert      OperationType = "delete_alert"
	OperationTypeExport           OperationType = "export"
	OperationTypeCreateFavorite   OperationType = "create_favorite"
	OperationTypeDeleteFavorite   OperationType = "delete_favorite"
	OperationTypeDeleteHistory    OperationType = "delete_history"
	OperationTypeChangePassword   OperationType = "change_password"
	OperationTypeCreateGroup      OperationType = "create_group"
	OperationTypeUpdateGroup      OperationType = "update_group"
	OperationTypeDeleteGroup      OperationType = "delete_group"
	OperationTypeAddMember        OperationType = "add_member"
	OperationTypeRemoveMember     OperationType = "remove_member"
	OperationTypeCreateFeedback   OperationType = "create_feedback"
)

// OperationLog represents an operation log entry.
type OperationLog struct {
	ID              uint            `gorm:"primaryKey;autoIncrement" json:"-"`
	SnowflakeID     int64           `gorm:"uniqueIndex;not null" json:"id,string"`
	UserID          *int64          `gorm:"default:null;index" json:"user_id"`
	OperationType   OperationType   `gorm:"type:varchar(50);not null;index" json:"operation_type"`
	OperationObject string          `gorm:"size:100" json:"operation_object,omitempty"`
	OperationDetail string          `gorm:"type:text" json:"operation_detail,omitempty"`
	OperationResult OperationResult `gorm:"type:varchar(20);not null" json:"operation_result"`
	IPAddress       string          `gorm:"size:50" json:"ip_address,omitempty"`
	CreatedAt       time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP;index" json:"created_at"`
}

// TableName returns the table name for OperationLog model.
func (OperationLog) TableName() string {
	return "operation_logs"
}
