// Package query implements the query history and favorite models.
package query

import (
	"time"
)

// Favorite represents a user's favorite query record.
type Favorite struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"-"`
	SnowflakeID   int64     `gorm:"uniqueIndex;not null" json:"id,string"`
	UserID        int64     `gorm:"not null;index" json:"user_id"`
	QueryRecordID *int64    `gorm:"index;uniqueIndex:idx_favorites_unique" json:"query_record_id,omitempty"`
	Name          string    `gorm:"type:varchar(200);not null" json:"name"`
	Input         string    `gorm:"type:text;not null" json:"input"`
	Sql           string    `gorm:"type:text;not null" json:"sql"`
	Description   string    `gorm:"type:text" json:"description,omitempty"`
	CreatedAt     time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName returns the table name for Favorite.
func (Favorite) TableName() string {
	return "favorites"
}
