package report

// CreateAlertRequest represents the request to create an alert rule.
type CreateAlertRequest struct {
	Name             string   `json:"name" validate:"required,min=3,max=100"`
	Description      string   `json:"description" validate:"max=200"`
	SQL              string   `json:"sql" validate:"required"`
	Condition        string   `json:"condition" validate:"required"`
	Recipients       []string `json:"recipients" validate:"required"`
	PushChannel      string   `json:"push_channel" validate:"omitempty,oneof=wechat email"`
	TriggerFrequency string   `json:"trigger_frequency" validate:"omitempty,oneof=every_time daily weekly"`
	SilenceStart     string   `json:"silence_start" validate:"omitempty"`
	SilenceEnd       string   `json:"silence_end" validate:"omitempty"`
	CreatedBy        int64    `json:"-" validate:"required"`
}

// UpdateAlertRequest represents the request to update an alert rule.
type UpdateAlertRequest struct {
	Name             string   `json:"name" validate:"omitempty,min=3,max=100"`
	Description      string   `json:"description" validate:"max=200"`
	SQL              string   `json:"sql"`
	Condition        string   `json:"condition"`
	Recipients       []string `json:"recipients"`
	PushChannel      string   `json:"push_channel" validate:"omitempty,oneof=wechat email"`
	TriggerFrequency string   `json:"trigger_frequency" validate:"omitempty,oneof=every_time daily weekly"`
	SilenceStart     string   `json:"silence_start" validate:"omitempty"`
	SilenceEnd       string   `json:"silence_end" validate:"omitempty"`
	Status           string   `json:"status" validate:"omitempty,oneof=active inactive"`
}

// ListAlertsRequest represents the request to list alert rules.
type ListAlertsRequest struct {
	Page     int    `form:"page" validate:"min=1"`
	PageSize int    `form:"page_size" validate:"min=1,max=100"`
	Name     string `form:"name" validate:"omitempty,max=100"`
	Status   string `form:"status" validate:"omitempty,oneof=active inactive"`
}

// AlertResponse represents the response for an alert rule.
type AlertResponse struct {
	ID               int64    `json:"id"`
	Name             string   `json:"name"`
	Description      string   `json:"description,omitempty"`
	SQL              string   `json:"sql"`
	Condition        string   `json:"condition"`
	Recipients       []string `json:"recipients"`
	PushChannel      string   `json:"push_channel"`
	TriggerFrequency string   `json:"trigger_frequency"`
	SilenceStart     string   `json:"silence_start,omitempty"`
	SilenceEnd       string   `json:"silence_end,omitempty"`
	Status           string   `json:"status"`
	LastTriggeredAt  string   `json:"last_triggered_at,omitempty"`
	CreatedBy        int64    `json:"created_by"`
	CreatedAt        string   `json:"created_at"`
}
