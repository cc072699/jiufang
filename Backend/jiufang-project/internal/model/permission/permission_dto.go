package permission

// PermissionRequest represents a single permission configuration.
type PermissionRequest struct {
	TableName       string `json:"table_name" binding:"required,min=1,max=100"`
	AllowedFields   string `json:"allowed_fields" binding:"required"` // JSON array format: ["id", "order_no", "amount"]
	FilterCondition string `json:"filter_condition"`                  // SQL condition: "status = 'approved'"
}

// PermissionResponse represents a permission in the response.
type PermissionResponse struct {
	ID              int64  `json:"id"`
	TableName       string `json:"table_name"`
	AllowedFields   string `json:"allowed_fields"`
	FilterCondition string `json:"filter_condition"`
}