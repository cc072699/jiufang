// Package dialog implements the dialog management model for multi-turn conversations.
// This file defines the DialogFavorite entity for user's favorite dialog sessions.
package dialog

import (
	"time"
)

// DialogFavorite represents a user's favorite dialog session.
// It allows users to bookmark important conversations for quick access.
type DialogFavorite struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"-"`
	SnowflakeID     int64     `gorm:"uniqueIndex;not null" json:"id,string"`
	UserID          int64     `gorm:"not null;index;uniqueIndex:uk_user_dialog" json:"user_id"`
	DialogSessionID int64     `gorm:"not null;index;uniqueIndex:uk_user_dialog" json:"dialog_session_id"`
	Title           string    `gorm:"type:varchar(100);default:null" json:"title,omitempty"`
	CreatedAt       time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName returns the table name for DialogFavorite.
func (DialogFavorite) TableName() string {
	return "dialog_favorites"
}

// DialogFavoriteCreate represents the request to create a new dialog favorite.
type DialogFavoriteCreate struct {
	UserID          int64
	DialogSessionID int64
	Title           string
}

// DialogFavoriteUpdate represents the request to update a dialog favorite.
type DialogFavoriteUpdate struct {
	Title string
}