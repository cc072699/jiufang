package user

import (
	"time"
)

// CreateUserRequest represents the request to create a user.
type CreateUserRequest struct {
	Username string   `json:"username" binding:"required,min=3,max=50"`
	Password string   `json:"password" binding:"required,min=8,max=100"`
	Email    string   `json:"email" binding:"required,email,max=100"`
	Role     Role     `json:"role" binding:"required,oneof=admin manager executive"`
	Groups   []string `json:"groups"` // User group IDs (snowflake IDs)
	Status   int      `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateUserRequest represents the request to update a user.
type UpdateUserRequest struct {
	Username string   `json:"username" binding:"omitempty,min=3,max=50"`
	Password string   `json:"password" binding:"omitempty,min=8,max=100"`
	Email    string   `json:"email" binding:"omitempty,email,max=100"`
	Role     Role     `json:"role" binding:"omitempty,oneof=admin manager executive"`
	Groups   []string `json:"groups"`
	Status   int      `json:"status" binding:"omitempty,oneof=0 1"`
}

// UserResponse represents a user in the response.
type UserResponse struct {
	ID        string    `json:"id"` // Snowflake ID as string to avoid JavaScript precision loss
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      Role      `json:"role"`
	Groups    []string  `json:"groups"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// ListUsersRequest represents the request to list users.
type ListUsersRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Username string `form:"username" binding:"max=50"`
	Role     Role   `form:"role" binding:"omitempty,oneof=admin manager executive"`
	Status   int    `form:"status" binding:"omitempty,oneof=0 1"`
}
