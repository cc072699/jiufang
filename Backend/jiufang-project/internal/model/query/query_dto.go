package query

import (
	"time"
)

// NaturalLanguageQueryRequest represents the request for natural language query.
type NaturalLanguageQueryRequest struct {
	Input              string `json:"input" binding:"required,min=1,max=500"`
	SessionID          string `json:"session_id"`          // Snowflake ID (19-digit number)
	ExecuteImmediately bool   `json:"execute_immediately"` // Default: true
}

// QueryResultResponse represents the response for a query result.
type QueryResultResponse struct {
	SessionID          string                   `json:"session_id"`
	QueryRecordID      string                   `json:"query_record_id"`
	Understanding      string                   `json:"understanding"`
	ResultType         string                   `json:"result_type"` // table/chart/empty
	SQL                string                   `json:"sql,omitempty"`
	Columns            []ColumnDefinition       `json:"columns,omitempty"`
	Rows               []map[string]interface{} `json:"rows,omitempty"`
	ChartConfig        interface{}              `json:"chart_config,omitempty"`
	SuggestedQuestions []string                 `json:"suggested_questions,omitempty"`
	CanExport          bool                     `json:"can_export"`
}

// ColumnDefinition represents a column in the query result.
type ColumnDefinition struct {
	Name string `json:"name"`
	Type string `json:"type"` // string/number/date
}

// QueryRecordResponse represents a query record in the response.
type QueryRecordResponse struct {
	ID            int64       `json:"id"`
	SessionID     string      `json:"session_id,omitempty"`
	Input         string      `json:"input"`
	SQL           string      `json:"sql,omitempty"`
	Status        QueryStatus `json:"status"`
	ErrorMessage  string      `json:"error_message,omitempty"`
	ResultCount   int         `json:"result_count,omitempty"`
	ExecutionTime int         `json:"execution_time,omitempty"` // milliseconds
	ResultData    string      `json:"result_data,omitempty"`    // JSON format
	CreatedAt     time.Time   `json:"created_at"`
}

// ListQueryRecordsRequest represents the request to list query records.
type ListQueryRecordsRequest struct {
	Page      int         `form:"page" binding:"min=1"`
	PageSize  int         `form:"page_size" binding:"min=1,max=100"`
	StartTime time.Time   `form:"start_time" format:"date-time"`
	EndTime   time.Time   `form:"end_time" format:"date-time"`
	Status    QueryStatus `form:"status" binding:"omitempty,oneof=success failed"`
}
