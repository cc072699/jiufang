package user

import (
	"time"
)

type UserGroupMember struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	SnowflakeID int64     `gorm:"uniqueIndex;not null" json:"snowflake_id,string"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	GroupID     uint      `gorm:"not null;index" json:"group_id"`
	CreatedAt   time.Time `json:"created_at"`
}

func (UserGroupMember) TableName() string {
	return "user_group_members"
}
