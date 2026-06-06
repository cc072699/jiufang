// Package feedback implements the user feedback model.
package feedback

import (
	"time"
)

// Rating represents the feedback rating type.
type Rating string

const (
	RatingSatisfied   Rating = "satisfied"
	RatingUnsatisfied Rating = "unsatisfied"
)

// Feedback represents a user feedback record.
type Feedback struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"-"`
	SnowflakeID   int64     `gorm:"uniqueIndex;not null" json:"id,string"`
	UserID        int64     `gorm:"not null;index" json:"user_id"`
	QueryRecordID int64     `gorm:"not null;index" json:"query_record_id"`
	QueryQuestion string    `gorm:"type:text;not null" json:"query_question"`
	Rating        Rating    `gorm:"type:varchar(20);not null" json:"rating"`
	Reason        string    `gorm:"type:text" json:"reason,omitempty"`
	CreatedAt     time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName returns the table name for Feedback model.
func (Feedback) TableName() string {
	return "feedbacks"
}
