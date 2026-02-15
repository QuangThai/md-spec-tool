package handlers

// ErrorResponse represents a standard error payload.
// Used by handlers that return errors directly; ErrorHandler middleware uses ErrorPayload.
type ErrorResponse struct {
	Error            string         `json:"error"`
	Code             string         `json:"code,omitempty"`
	ValidationReason string         `json:"validation_reason,omitempty"`
	Details          map[string]any `json:"details,omitempty"`
	RequestID        string         `json:"request_id,omitempty"`
}
