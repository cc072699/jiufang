package report

import "time"

// CreateReportRequest represents the request to create a scheduled report.
type CreateReportRequest struct {
	Name         string       `json:"name" binding:"required,min=1,max=100"`
	SQL          string       `json:"sql" binding:"required,min=1,max=5000"`
	ScheduleType ScheduleType `json:"schedule_type" binding:"required,oneof=daily weekly monthly"`
	ScheduleTime string       `json:"schedule_time" binding:"required"`
	Recipients   []string     `json:"recipients" binding:"required,min=1"`
	PushChannel  PushChannel  `json:"push_channel" binding:"omitempty,oneof=wechat email"`
	Description  string       `json:"description" binding:"max=200"`
	CreatedBy    int64        `json:"-"`
}

// UpdateReportRequest represents the request to update a scheduled report.
type UpdateReportRequest struct {
	Name         string       `json:"name" binding:"min=1,max=100"`
	SQL          string       `json:"sql" binding:"min=1,max=5000"`
	ScheduleType ScheduleType `json:"schedule_type" binding:"oneof=daily weekly monthly"`
	ScheduleTime string       `json:"schedule_time"`
	Recipients   []string     `json:"recipients"`
	PushChannel  PushChannel  `json:"push_channel" binding:"omitempty,oneof=wechat email"`
	Description  string       `json:"description" binding:"max=200"`
}

// ReportResponse represents the response for a scheduled report.
type ReportResponse struct {
	ID           int64        `json:"id"`
	Name         string       `json:"name"`
	SQL          string       `json:"sql"`
	ScheduleType ScheduleType `json:"schedule_type"`
	ScheduleTime string       `json:"schedule_time"`
	Recipients   []string     `json:"recipients"`
	PushChannel  PushChannel  `json:"push_channel"`
	Description  string       `json:"description"`
	Status       ReportStatus `json:"status"`
	CreatedBy    int64        `json:"created_by"`
	CreatedAt    string       `json:"created_at"`
	UpdatedAt    string       `json:"updated_at,omitempty"`
}

// ListReportsRequest represents the request to list scheduled reports.
type ListReportsRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Name     string `form:"name"`
	Status   string `form:"status" binding:"omitempty,oneof=active inactive"`
}

// GetPushRecordsRequest represents the request to get push records.
type GetPushRecordsRequest struct {
	Page      int        `form:"page" binding:"min=1"`
	PageSize  int        `form:"page_size" binding:"min=1,max=100"`
	PushType  PushType   `form:"push_type" binding:"omitempty,oneof=report alert"`
	Status    PushStatus `form:"status" binding:"omitempty,oneof=success failed retrying"`
	StartTime time.Time  `form:"start_time" format:"date-time"`
	EndTime   time.Time  `form:"end_time" format:"date-time"`
}
