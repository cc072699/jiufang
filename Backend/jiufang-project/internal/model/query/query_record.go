// Package query implements the query history and favorite models.
package query

import (
	"time"
)

// QueryStatus represents the status of a query execution.
type QueryStatus string

const (
	QueryStatusSuccess QueryStatus = "success"
	QueryStatusFailed  QueryStatus = "failed"
)

// SessionStatus represents the status of a query session.
type SessionStatus string

const (
	SessionStatusActive SessionStatus = "active"
	SessionStatusClosed SessionStatus = "closed"
)

// QueryRecord represents a single query execution record.
type QueryRecord struct {
	ID            uint        `gorm:"primaryKey;autoIncrement" json:"-"`
	SnowflakeID   int64       `gorm:"uniqueIndex;not null" json:"id,string"`
	SessionID     int64       `gorm:"not null;index" json:"session_id"`
	UserID        int64       `gorm:"not null;index" json:"user_id"`
	Input         string      `gorm:"type:text;not null" json:"input"`
	SQL           string      `gorm:"type:text;not null" json:"sql"`
	Status        QueryStatus `gorm:"type:varchar(20);not null" json:"status"`
	ErrorMessage  string      `gorm:"type:text" json:"error_message,omitempty"`
	ResultCount   int         `gorm:"default:null" json:"result_count,omitempty"`
	ExecutionTime int         `gorm:"default:null" json:"execution_time,omitempty"` // milliseconds
	ResultData    string      `gorm:"type:text" json:"result_data,omitempty"`
	CreatedAt     time.Time   `gorm:"not null;default:CURRENT_TIMESTAMP;index" json:"created_at"`
}

// TableName returns the table name for QueryRecord.
func (QueryRecord) TableName() string {
	return "query_records"
}

// IsSuccess returns true if the query execution was successful.
func (q *QueryRecord) IsSuccess() bool {
	return q.Status == QueryStatusSuccess
}

// GetSummary returns a summary of the query record for display.
func (q *QueryRecord) GetSummary() string {
	if len(q.Input) > 20 {
		return q.Input[:20] + "..."
	}
	return q.Input
}