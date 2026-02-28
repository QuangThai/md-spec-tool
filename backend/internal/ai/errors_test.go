package ai

import (
	"context"
	"errors"
	"testing"
)

func TestClassifyError_Transient(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		err        error
	}{
		{"rate_limited", 429, errors.New("rate limited")},
		{"server_error_500", 500, errors.New("internal server error")},
		{"server_error_502", 502, errors.New("bad gateway")},
		{"server_error_503", 503, errors.New("service unavailable")},
		{"timeout", 0, context.DeadlineExceeded},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classified := ClassifyError(tt.statusCode, tt.err)
			if classified.Category != ErrorCategoryTransient {
				t.Errorf("expected Transient, got %s", classified.Category)
			}
			if !classified.ShouldRetry {
				t.Error("transient errors should be retryable")
			}
		})
	}
}

func TestClassifyError_Permanent(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"bad_request", 400},
		{"unauthorized", 401},
		{"forbidden", 403},
		{"not_found", 404},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classified := ClassifyError(tt.statusCode, errors.New("error"))
			if classified.Category != ErrorCategoryPermanent {
				t.Errorf("expected Permanent, got %s", classified.Category)
			}
			if classified.ShouldRetry {
				t.Error("permanent errors should NOT be retryable")
			}
		})
	}
}

func TestClassifyError_Content(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"refusal", ErrAIRefused},
		{"truncated", ErrAITruncated},
		{"content_filter", ErrAIContentFiltered},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classified := ClassifyError(0, tt.err)
			if classified.Category != ErrorCategoryContent {
				t.Errorf("expected Content, got %s", classified.Category)
			}
		})
	}
}

func TestClassifyError_UnknownDefaultsToTransient(t *testing.T) {
	classified := ClassifyError(0, errors.New("some random error"))
	if classified.Category != ErrorCategoryTransient {
		t.Errorf("expected Transient for unknown error, got %s", classified.Category)
	}
	if !classified.ShouldRetry {
		t.Error("unknown errors should default to retryable")
	}
}
