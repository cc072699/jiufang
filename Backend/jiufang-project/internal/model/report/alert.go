// Package report implements the alert rule model.
package report

import (
	"time"
)

// AlertStatus represents the status of an alert rule.
type AlertStatus string

const (
	AlertStatusActive   AlertStatus = "active"
	AlertStatusInactive AlertStatus = "inactive"
)

// TriggerFrequency represents the trigger frequency of an alert.
type TriggerFrequency string

const (
	TriggerFrequencyEveryTime TriggerFrequency = "every_time"
	TriggerFrequencyDaily     TriggerFrequency = "daily"
	TriggerFrequencyWeekly    TriggerFrequency = "weekly"
)

// Alert represents an alert rule configuration.
type Alert struct {
	ID               uint             `gorm:"primaryKey;autoIncrement" json:"-"`
	SnowflakeID      int64            `gorm:"uniqueIndex;not null" json:"id,string"`
	Name             string           `gorm:"size:100;not null" json:"name"`
	Description      string           `gorm:"size:200" json:"description,omitempty"`
	SQL              string           `gorm:"type:text;not null" json:"sql"`
	Condition        string           `gorm:"size:200;not null" json:"condition"`
	Recipients       string           `gorm:"type:text;not null" json:"recipients"` // JSON array of user IDs or emails
	PushChannel      PushChannel      `gorm:"type:varchar(20);not null;default:wechat" json:"push_channel"`
	TriggerFrequency TriggerFrequency `gorm:"type:varchar(20);not null;default:every_time" json:"trigger_frequency"`
	SilenceStart     *time.Time       `gorm:"type:time" json:"silence_start,omitempty"`
	SilenceEnd       *time.Time       `gorm:"type:time" json:"silence_end,omitempty"`
	Status           AlertStatus      `gorm:"type:varchar(20);not null;default:active" json:"status"`
	LastTriggeredAt  *time.Time       `gorm:"default:null" json:"last_triggered_at,omitempty"`
	CreatedBy        int64            `gorm:"not null;index" json:"created_by"`
	CreatedAt        time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName returns the table name for Alert model.
func (Alert) TableName() string {
	return "alerts"
}
