// Package feedback implements the user feedback DTOs.
package feedback

// CreateFeedbackRequest represents the request to create a feedback.
type CreateFeedbackRequest struct {
	QueryRecordID int64  `json:"query_record_id" validate:"required"`
	Rating        string `json:"rating" validate:"required,oneof=satisfied unsatisfied"`
	Reason        string `json:"reason" validate:"omitempty,max=500"`
}

// FeedbackResponse represents the response for a feedback.
type FeedbackResponse struct {
	ID            int64  `json:"id"`
	UserID        int64  `json:"user_id"`
	QueryRecordID int64  `json:"query_record_id"`
	QueryQuestion string `json:"query_question"`
	Rating        string `json:"rating"`
	Reason        string `json:"reason,omitempty"`
	CreatedAt     string `json:"created_at"`
}

// ListFeedbacksRequest represents the request to list feedbacks.
type ListFeedbacksRequest struct {
	Page    int    `form:"page" validate:"min=1"`
	Size    int    `form:"size" validate:"min=1,max=100"`
	UserID  int64  `form:"user_id" validate:"omitempty"`
	Rating  string `form:"rating" validate:"omitempty,oneof=satisfied unsatisfied"`
}