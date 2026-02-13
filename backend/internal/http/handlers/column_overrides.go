package handlers

import (
	"encoding/json"
	"fmt"
	"strings"
)

func parseColumnOverrides(raw string) (map[string]string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	var overrides map[string]string
	if err := json.Unmarshal([]byte(trimmed), &overrides); err != nil {
		return nil, fmt.Errorf("invalid column_overrides JSON")
	}

	cleaned := make(map[string]string)
	for header, field := range overrides {
		header = strings.TrimSpace(header)
		field = strings.TrimSpace(field)
		if header == "" || field == "" {
			continue
		}
		cleaned[header] = field
	}

	if len(cleaned) == 0 {
		return nil, nil
	}

	return cleaned, nil
}
