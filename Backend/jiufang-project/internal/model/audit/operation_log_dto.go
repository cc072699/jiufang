// Package audit implements the operation log DTOs.
package audit

// ListOperationLogsRequest represents the request to list operation logs.
type ListOperationLogsRequest struct {
	Page           int    `form:"page" validate:"min=1"`
PageSize       int    `form:"page_size" validate:"min=1,max=100"`
	UserID         int64  `form:"user_id" validate:"omitempty"`
	OperationType  string `form:"operation_type" validate:"omitempty"`
StartTime      string `form:"start_time" validate:"omitempty"`
EndTime        string `form:"end_time" validate:"omitempty"`
}

// OperationLogResponse represents the response for an operation log.
type OperationLogResponse struct {
	ID              int64  `json:"id"`
	UserID          int64  `json:"user_id,omitempty"`
Username        string `json:"username"`
	OperationType   string `json:"operation_type"`
	OperationObject string `json:"operation_object,omitempty"`
	OperationDetail string `json:"operation_detail,omitempty"`
	OperationResult string `json:"operation_result"`
	IPAddress       string `json:"ip_address,omitempty"`
	CreatedAt       string `json:"created_at"`
}