// Package report implements the scheduled report and push record models.
package report

import (
	"time"

	"gorm.io/gorm"
)

// ReportStatus represents the status of a scheduled report.
type ReportStatus string

const (
	ReportStatusActive   ReportStatus = "active"
	ReportStatusInactive ReportStatus = "inactive"
)

// ScheduleType represents the type of schedule.
type ScheduleType string

const (
	ScheduleTypeDaily   ScheduleType = "daily"
	ScheduleTypeWeekly  ScheduleType = "weekly"
	ScheduleTypeMonthly ScheduleType = "monthly"
)

// PushChannel represents the push channel type.
type PushChannel string

const (
	PushChannelWeChat PushChannel = "wechat"
	PushChannelEmail  PushChannel = "email"
)

// PushType represents the push type.
type PushType string

const (
	PushTypeReport PushType = "report"
	PushTypeAlert  PushType = "alert"
)

// PushStatus represents the status of a push record.
type PushStatus string

const (
	PushStatusSuccess  PushStatus = "success"
	PushStatusFailed   PushStatus = "failed"
	PushStatusRetrying PushStatus = "retrying"
)

// ScheduledReport represents a scheduled report configuration.
type ScheduledReport struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"-"`
	SnowflakeID  int64          `gorm:"uniqueIndex;not null" json:"id,string"`
	Name         string         `gorm:"size:100;not null" json:"name"`
	Description  string         `gorm:"size:200" json:"description,omitempty"`
	SQL          string         `gorm:"type:text;not null" json:"sql"`         // SQL语句
	ScheduleType ScheduleType   `gorm:"size:20;not null" json:"schedule_type"` // 定时类型（daily/weekly/monthly）
	ScheduleTime string         `gorm:"size:50;not null" json:"schedule_time"` // 定时时间（ISO8601格式或时间表达式）
	Recipients   string         `gorm:"type:text;not null" json:"recipients"`  // 接收者列表（JSON数组格式）
	Status       ReportStatus   `gorm:"size:20;not null;default:active" json:"status"`
	CreatedBy    int64          `gorm:"not null;index" json:"created_by"`
	CreatedAt    time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name for ScheduledReport.
func (ScheduledReport) TableName() string {
	return "scheduled_reports"
}

// IsActive returns true if the report is active.
func (r *ScheduledReport) IsActive() bool {
	return r.Status == ReportStatusActive
}
