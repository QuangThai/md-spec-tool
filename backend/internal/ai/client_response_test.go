package ai

import (
	"errors"
	"testing"
)

func TestHandleFinishReason_Stop(t *testing.T) {
	err := handleFinishReason("stop")
	if err != nil {
		t.Errorf("expected no error for stop, got: %v", err)
	}
}

func TestHandleFinishReason_Length(t *testing.T) {
	err := handleFinishReason("length")
	if !errors.Is(err, ErrAITruncated) {
		t.Errorf("expected ErrAITruncated for length, got: %v", err)
	}
}

func TestHandleFinishReason_ContentFilter(t *testing.T) {
	err := handleFinishReason("content_filter")
	if !errors.Is(err, ErrAIContentFiltered) {
		t.Errorf("expected ErrAIContentFiltered, got: %v", err)
	}
}

func TestHandleRefusal_NoRefusal(t *testing.T) {
	err := handleRefusal("")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestHandleRefusal_WithRefusal(t *testing.T) {
	err := handleRefusal("I cannot process this content")
	if !errors.Is(err, ErrAIRefused) {
		t.Errorf("expected ErrAIRefused, got: %v", err)
	}
}
