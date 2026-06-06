// Package report implements the scheduled report and push record models.
package report

import (
	"time"
)

// PushRecord represents a push execution record.
type PushRecord struct {
	ID           uint        `gorm:"primaryKey;autoIncrement" json:"-"`
	SnowflakeID  int64       `gorm:"uniqueIndex;not null" json:"id,string"`
	ReportID     int64       `gorm:"index" json:"report_id,omitempty"`
	AlertRuleID  int64       `gorm:"index" json:"alert_rule_id,omitempty"`
	PushType     PushType    `gorm:"type:varchar(20);not null" json:"push_type"`
	PushContent  string      `gorm:"type:text;not null" json:"push_content"` // Markdown format
	PushTargets  string      `gorm:"type:text;not null" json:"push_targets"` // JSON array
	PushChannel  PushChannel `gorm:"type:varchar(20);not null" json:"push_channel"`
	PushStatus   PushStatus  `gorm:"type:varchar(20);not null" json:"push_status"`
	PushTime     time.Time   `gorm:"not null;default:CURRENT_TIMESTAMP;index" json:"push_time"`
	ErrorMessage string      `gorm:"size:500" json:"error_message,omitempty"`
	RetryCount   int         `gorm:"not null;default:0" json:"retry_count"`
}

// TableName returns the table name for PushRecord.
func (PushRecord) TableName() string {
	return "push_records"
}

// IsSuccess returns true if the push was successful.
func (p *PushRecord) IsSuccess() bool {
	return p.PushStatus == PushStatusSuccess
}

// CanRetry returns true if the push can be retried.
func (p *PushRecord) CanRetry() bool {
	return p.PushStatus == PushStatusRetrying && p.RetryCount < 3
}
