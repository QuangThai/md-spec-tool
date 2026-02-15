package gsheetutils

import (
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// ParseGoogleSheetURL extracts the sheet ID from various Google Sheets URL formats
// Returns sheetID, gid, and a boolean indicating success
func ParseGoogleSheetURL(urlStr string) (sheetID string, gid string, ok bool) {
	// Format 1: https://docs.google.com/spreadsheets/d/SHEET_ID/edit#gid=GID
	// Format 2: https://docs.google.com/spreadsheets/d/SHEET_ID/edit
	// Format 3: https://docs.google.com/spreadsheets/d/SHEET_ID

	// Parse URL properly
	u, err := url.Parse(urlStr)
	if err != nil {
		slog.Warn("Invalid URL format", "url", urlStr, "error", err)
		return "", "", false
	}

	// Verify host is a supported Google Sheets domain
	host := strings.ToLower(u.Host)
	if host != "docs.google.com" && host != "spreadsheets.google.com" {
		slog.Warn("Not a Google Docs URL", "host", u.Host)
		return "", "", false
	}

	// Path should contain /spreadsheets/d/{id}
	// Use regex to extract SHEET_ID pattern (alphanumeric, hyphens, underscores)
	sheetIDPattern := regexp.MustCompile(`/spreadsheets/d/([a-zA-Z0-9\-_]+)`)
	matches := sheetIDPattern.FindStringSubmatch(u.Path)

	if len(matches) < 2 {
		slog.Warn("Sheet ID not found in URL path", "path", u.Path)
		return "", "", false
	}

	sheetID = matches[1]

	// Validate sheet ID length (Google Sheet IDs are typically 40+ chars but vary)
	if len(sheetID) == 0 {
		return "", "", false
	}

	// Extract gid from fragment or query parameter
	// Fragment takes precedence (e.g., #gid=123)
	if u.Fragment != "" {
		gidPattern := regexp.MustCompile(`gid=(\d+)`)
		gidMatches := gidPattern.FindStringSubmatch(u.Fragment)
		if len(gidMatches) >= 2 {
			gid = gidMatches[1]
		}
	}

	// If not found in fragment, check query parameters
	if gid == "" {
		gid = u.Query().Get("gid")
	}

	return sheetID, gid, true
}

// SelectGID chooses between request-provided GID and URL-extracted GID
func SelectGID(requestGID string, urlGID string) string {
	requestGID = strings.TrimSpace(requestGID)
	if requestGID != "" {
		return requestGID
	}
	return urlGID
}

// ValidateGID validates that GID is numeric if provided
func ValidateGID(gid string) error {
	if gid == "" {
		return nil
	}
	if _, err := strconv.ParseInt(gid, 10, 64); err != nil {
		return fmt.Errorf("gid must be numeric")
	}
	return nil
}
