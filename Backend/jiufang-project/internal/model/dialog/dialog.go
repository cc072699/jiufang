// Package dialog implements the dialog management model for multi-turn conversations.
// This package manages dialog sessions, context, and conversation history.
package dialog

import (
	"time"
)

// DialogSession represents a dialog session entity.
// It stores the metadata of a conversation session (context is stored in Redis).
type DialogSession struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	SnowflakeID    string     `gorm:"uniqueIndex;size:19;not null" json:"snowflake_id"`
	UserID         uint       `gorm:"index;not null" json:"user_id"`
	QuerySessionID *uint      `gorm:"index" json:"query_session_id,omitempty"`
	Status         string     `gorm:"size:20;default:active;not null" json:"status"` // active/closed
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	ClosedAt       *time.Time `json:"closed_at,omitempty"`
}

// TableName returns the table name for DialogSession.
func (DialogSession) TableName() string {
	return "dialog_sessions"
}

// IsActive checks if the dialog session is active.
func (d *DialogSession) IsActive() bool {
	return d.Status == "active"
}

// Close closes the dialog session.
func (d *DialogSession) Close() {
	d.Status = "closed"
	now := time.Now()
	d.ClosedAt = &now
	d.UpdatedAt = now
}

// SessionStatus represents the status of a dialog session.
type SessionStatus string

const (
	StatusActive SessionStatus = "active"
	StatusClosed SessionStatus = "closed"
)

// String returns the string representation of the status.
func (s SessionStatus) String() string {
	return string(s)
}

// DialogSessionCreate represents the request to create a new dialog session.
type DialogSessionCreate struct {
	UserID         uint
	QuerySessionID *uint
}

// DialogSessionUpdate represents the request to update a dialog session.
type DialogSessionUpdate struct {
	Status         *string
	QuerySessionID *uint
}
