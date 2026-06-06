// Package agent defines the data models for AI Agent module.
package agent

import "fmt"

// EntityType represents the type of extracted entity from natural language.
// Based on PRD BR-002, entities include time range, amount, document type, supplier, etc.
type EntityType string

const (
	// EntityTypeTimeRange represents time range entity (e.g., "last month", "this week", "2024-01-01 to 2024-12-31")
	EntityTypeTimeRange EntityType = "time_range"

	// EntityTypeAmount represents amount/money condition (e.g., "more than 10000", "between 5000 and 10000")
	EntityTypeAmount EntityType = "amount"

	// EntityTypeDocumentType represents document type (e.g., "purchase order", "sales order", "payment")
	EntityTypeDocumentType EntityType = "document_type"

	// EntityTypeSupplier represents supplier entity (e.g., "supplier A", "company B")
	EntityTypeSupplier EntityType = "supplier"

	// EntityTypeDepartment represents department entity (e.g., "sales department", "finance department")
	EntityTypeDepartment EntityType = "department"

	// EntityTypeCustomer represents customer entity (e.g., "customer X", "client Y")
	EntityTypeCustomer EntityType = "customer"

	// EntityTypeProduct represents product entity (e.g., "product A", "item B")
	EntityTypeProduct EntityType = "product"

	// EntityTypeStatus represents status entity (e.g., "completed", "pending", "approved")
	EntityTypeStatus EntityType = "status"

	// EntityTypeGroupBy represents grouping dimension (e.g., "by supplier", "by month")
	EntityTypeGroupBy EntityType = "group_by"

	// EntityTypeOrderBy represents sorting condition (e.g., "by amount descending", "by date ascending")
	EntityTypeOrderBy EntityType = "order_by"

	// EntityTypeLimit represents result limit (e.g., "top 10", "first 20")
	EntityTypeLimit EntityType = "limit"
)

// Entity represents an extracted entity from natural language input.
// Each entity has a type, raw text, and normalized value.
type Entity struct {
	// Type is the entity type
	Type EntityType `json:"type"`

	// Value is the normalized/standardized value for SQL generation
	// e.g., "last month" -> "2024-05-01 to 2024-05-31"
	Value string `json:"value"`

	// RawText is the original text from user input
	RawText string `json:"raw_text"`

	// Normalized is the intermediate normalized form
	// e.g., "上个月" -> "last_month"
	Normalized string `json:"normalized"`

	// Confidence is the confidence score of entity extraction (0.0 ~ 1.0)
	Confidence float64 `json:"confidence"`

	// StartPos is the start position in the original text
	StartPos int `json:"start_pos"`

	// EndPos is the end position in the original text
	EndPos int `json:"end_pos"`
}

// TimeRangeValue represents a parsed time range with start and end dates.
type TimeRangeValue struct {
	StartDate  string `json:"start_date"`  // Format: YYYY-MM-DD
	EndDate    string `json:"end_date"`    // Format: YYYY-MM-DD
	IsRelative bool   `json:"is_relative"` // true if relative time (e.g., "last month")
}

// AmountValue represents a parsed amount condition with operator and value.
type AmountValue struct {
	Operator string  `json:"operator"` // e.g., ">", "<", "=", "between"
	Value    float64 `json:"value"`
	ValueEnd float64 `json:"value_end"` // For "between" operator
}

// GroupByValue represents a parsed grouping dimension.
type GroupByValue struct {
	Field     string `json:"field"`     // e.g., "supplier_id", "month"
	Direction string `json:"direction"` // For time-based grouping: "asc", "desc"
}

// OrderByValue represents a parsed sorting condition.
type OrderByValue struct {
	Field     string `json:"field"`     // e.g., "amount", "created_at"
	Direction string `json:"direction"` // "asc" or "desc"
}

// LimitValue represents a parsed result limit.
type LimitValue struct {
	Value int `json:"value"` // e.g., 10, 20
}

// EntityCollection holds all extracted entities for a query.
type EntityCollection struct {
	TimeRange    *TimeRangeValue `json:"time_range,omitempty"`
	Amount       *AmountValue    `json:"amount,omitempty"`
	DocumentType string          `json:"document_type,omitempty"`
	Supplier     string          `json:"supplier,omitempty"`
	Department   string          `json:"department,omitempty"`
	Customer     string          `json:"customer,omitempty"`
	Product      string          `json:"product,omitempty"`
	Status       string          `json:"status,omitempty"`
	GroupBy      []GroupByValue  `json:"group_by,omitempty"`
	OrderBy      []OrderByValue  `json:"order_by,omitempty"`
	Limit        *LimitValue     `json:"limit,omitempty"`
}

// ToEntityList converts EntityCollection to a list of Entity.
func (ec *EntityCollection) ToEntityList() []Entity {
	entities := []Entity{}

	if ec.TimeRange != nil {
		entities = append(entities, Entity{
			Type:       EntityTypeTimeRange,
			Value:      ec.TimeRange.StartDate + " to " + ec.TimeRange.EndDate,
			Normalized: "time_range",
		})
	}

	if ec.Amount != nil {
		entities = append(entities, Entity{
			Type:       EntityTypeAmount,
			Value:      ec.Amount.Operator + " " + fmt.Sprintf("%.2f", ec.Amount.Value),
			Normalized: "amount",
		})
	}

	if ec.DocumentType != "" {
		entities = append(entities, Entity{
			Type:       EntityTypeDocumentType,
			Value:      ec.DocumentType,
			Normalized: "document_type",
		})
	}

	return entities
}
