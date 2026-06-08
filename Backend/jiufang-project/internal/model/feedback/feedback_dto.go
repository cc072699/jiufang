// Package feedback implements the user feedback DTOs.
package feedback

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// CreateFeedbackRequest represents the request to create a feedback.
type CreateFeedbackRequest struct {
	QueryRecordID int64  `json:"query_record_id" validate:"required"`
	Rating        string `json:"rating" validate:"required,oneof=satisfied unsatisfied"`
	Reason        string `json:"reason" validate:"omitempty,max=500"`
}

// UnmarshalJSON handles both string and number formats for query_record_id.
func (r *CreateFeedbackRequest) UnmarshalJSON(data []byte) error {
	type Alias CreateFeedbackRequest
	aux := &struct {
		QueryRecordID json.RawMessage `json:"query_record_id"`
		*Alias
	}{Alias: (*Alias)(r)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if len(aux.QueryRecordID) > 0 {
		// Try number first
		var num int64
		if err := json.Unmarshal(aux.QueryRecordID, &num); err == nil {
			r.QueryRecordID = num
			return nil
		}
		// Try string
		var str string
		if err := json.Unmarshal(aux.QueryRecordID, &str); err == nil {
			parsed, err := strconv.ParseInt(str, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid query_record_id: %s", str)
			}
			r.QueryRecordID = parsed
			return nil
		}
		return fmt.Errorf("query_record_id must be a number or numeric string")
	}
	return nil
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