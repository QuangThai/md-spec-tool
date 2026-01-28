package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type ActivityLog struct {
	ID             string          `db:"id" json:"id"`
	SpecID         string          `db:"spec_id" json:"spec_id"`
	UserID         string          `db:"user_id" json:"user_id"`
	Username       string          `json:"username"` // From JOIN
	Action         string          `db:"action" json:"action"`
	ResourceType   string          `db:"resource_type" json:"resource_type"`
	ResourceID     *string         `db:"resource_id" json:"resource_id"`
	Details        json.RawMessage `db:"details" json:"details"`
	FiltersApplied json.RawMessage `db:"filters_applied" json:"filters_applied"` // NEW: Audit trail
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
}

// ActivityDetails stores change information
type ActivityDetails struct {
	Field    string `json:"field,omitempty"`
	OldValue string `json:"old_value,omitempty"`
	NewValue string `json:"new_value,omitempty"`
}

// Scan implements sql.Scanner interface
func (a *ActivityDetails) Scan(value interface{}) error {
	bytes, _ := value.([]byte)
	return json.Unmarshal(bytes, &a)
}

// Value implements driver.Valuer interface
func (a ActivityDetails) Value() (driver.Value, error) {
	return json.Marshal(a)
}

type ActivityLogResponse struct {
	Logs  []*ActivityLog `json:"logs"`
	Count int            `json:"count"`
}

// NEW: Activity filtering

type ActivityFilterRequest struct {
	SpecID       string     `json:"spec_id" binding:"required"`
	Actions      []string   `json:"actions"`       // Filter: created, updated, shared, commented, deleted
	UserID       string     `json:"user_id"`       // Filter: specific user
	ResourceType string     `json:"resource_type"` // Filter: spec, comment, share
	StartDate    *time.Time `json:"start_date"`    // Filter: date range
	EndDate      *time.Time `json:"end_date"`      // Filter: date range
	Limit        int        `json:"limit" binding:"min=1,max=1000"`
	Offset       int        `json:"offset" binding:"min=0"`
}

type ActivityFilterResponse struct {
	Data       []*ActivityLog `json:"data"`
	Pagination PaginationInfo `json:"pagination"`
}

type PaginationInfo struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

type ActivityStatsResponse struct {
	Created   int `json:"created"`
	Updated   int `json:"updated"`
	Shared    int `json:"shared"`
	Commented int `json:"commented"`
	Deleted   int `json:"deleted"`
}

type ActivityExportRequest struct {
	SpecID  string                  `json:"spec_id" binding:"required"`
	Filters ActivityFilterRequest   `json:"filters" binding:"required"`
	Format  string                  `json:"format" binding:"required,oneof=json csv"`
}
