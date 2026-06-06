// Package query implements the query history and favorite models.
package query

import (
	"time"
)

// QuerySession represents a query conversation session.
type QuerySession struct {
	ID          uint          `gorm:"primaryKey;autoIncrement" json:"-"`
	SnowflakeID int64         `gorm:"uniqueIndex;not null" json:"id,string"`
	UserID      int64         `gorm:"not null;index" json:"user_id"`
	DialogID    int64         `gorm:"default:null" json:"dialog_id,omitempty"`
	Status      SessionStatus `gorm:"type:varchar(20);not null;default:active;index" json:"status"`
	CreatedAt   time.Time     `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time     `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	ClosedAt    time.Time     `gorm:"default:null" json:"closed_at,omitempty"`
}

// TableName returns the table name for QuerySession.
func (QuerySession) TableName() string {
	return "query_sessions"
}

// IsActive returns true if the session is active.
func (s *QuerySession) IsActive() bool {
	return s.Status == SessionStatusActive
}

// Close closes the session.
func (s *QuerySession) Close() {
	s.Status = SessionStatusClosed
	s.ClosedAt = time.Now()
	s.UpdatedAt = time.Now()
}