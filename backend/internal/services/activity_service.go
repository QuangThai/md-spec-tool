package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/yourorg/md-spec-tool/internal/models"
	"github.com/yourorg/md-spec-tool/internal/repositories"
)

type ActivityService struct {
	activityRepo     *repositories.ActivityRepository
	notificationRepo *repositories.NotificationRepository
	userRepo         *repositories.UserRepository
}

func NewActivityService(activityRepo *repositories.ActivityRepository, notificationRepo *repositories.NotificationRepository, userRepo *repositories.UserRepository) *ActivityService {
	return &ActivityService{
		activityRepo:     activityRepo,
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
	}
}

// LogSpecCreated logs spec creation
func (s *ActivityService) LogSpecCreated(ctx context.Context, specID, userID, title string) error {
	details := &models.ActivityDetails{
		Field:    "title",
		NewValue: title,
	}
	return s.activityRepo.Log(ctx, specID, userID, "created", "spec", nil, details)
}

// LogSpecUpdated logs spec modification
func (s *ActivityService) LogSpecUpdated(ctx context.Context, specID, userID, field, oldValue, newValue string) error {
	details := &models.ActivityDetails{
		Field:    field,
		OldValue: oldValue,
		NewValue: newValue,
	}
	return s.activityRepo.Log(ctx, specID, userID, "updated", "spec", nil, details)
}

// LogSpecShared logs spec sharing
func (s *ActivityService) LogSpecShared(ctx context.Context, specID, userID, shareID, sharedWithUserID string) error {
	details := &models.ActivityDetails{
		Field:    "permission",
		NewValue: "shared",
	}
	return s.activityRepo.Log(ctx, specID, userID, "shared", "share", &shareID, details)
}

// LogSpecUnshared logs spec unsharing
func (s *ActivityService) LogSpecUnshared(ctx context.Context, specID, userID, sharedWithUserID string) error {
	details := &models.ActivityDetails{
		Field:    "permission",
		OldValue: "shared",
	}
	return s.activityRepo.Log(ctx, specID, userID, "unshared", "share", nil, details)
}

// LogCommentAdded logs comment creation
func (s *ActivityService) LogCommentAdded(ctx context.Context, specID, userID, commentID string) error {
	return s.activityRepo.Log(ctx, specID, userID, "commented", "comment", &commentID, nil)
}

// LogSpecDeleted logs soft delete
func (s *ActivityService) LogSpecDeleted(ctx context.Context, specID, userID string) error {
	return s.activityRepo.Log(ctx, specID, userID, "deleted", "spec", nil, nil)
}

// GetSpecActivity retrieves activity for a spec
func (s *ActivityService) GetSpecActivity(ctx context.Context, specID string, limit, offset int) ([]*models.ActivityLog, int, error) {
	return s.activityRepo.GetBySpecID(ctx, specID, limit, offset)
}

// GetUserActivity retrieves activity for a user
func (s *ActivityService) GetUserActivity(ctx context.Context, userID string, limit, offset int) ([]*models.ActivityLog, int, error) {
	return s.activityRepo.GetByUserID(ctx, userID, limit, offset)
}

// ExtractMentions extracts @mentions from content
func (s *ActivityService) ExtractMentions(content string) []string {
	// Match @username patterns
	re := regexp.MustCompile(`@([a-zA-Z0-9._-]+)`)
	matches := re.FindAllStringSubmatch(content, -1)

	var usernames []string
	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			usernames = append(usernames, match[1])
			seen[match[1]] = true
		}
	}
	return usernames
}

// ResolveUsersByEmail gets user IDs from emails (username format)
func (s *ActivityService) ResolveUsersByEmail(ctx context.Context, emails []string) ([]string, error) {
	var userIDs []string
	for _, email := range emails {
		user, err := s.userRepo.GetByEmail(ctx, email)
		if err == nil && user != nil {
			userIDs = append(userIDs, user.ID)
		}
	}
	return userIDs, nil
}

// PHASE 7: Activity filtering and export

// GetFilteredActivities returns activities with filters applied
func (s *ActivityService) GetFilteredActivities(ctx context.Context, filters *models.ActivityFilterRequest) ([]*models.ActivityLog, int, error) {
	return s.activityRepo.GetActivitiesWithFilters(ctx, filters)
}

// GetActivityStats returns statistics about activities
func (s *ActivityService) GetActivityStats(ctx context.Context, specID string) (*models.ActivityStatsResponse, error) {
	return s.activityRepo.GetActivityStats(ctx, specID)
}

// ExportActivities exports activities as JSON or CSV
func (s *ActivityService) ExportActivities(ctx context.Context, filters *models.ActivityFilterRequest, format string) ([]byte, string, error) {
	logs, _, err := s.activityRepo.GetActivitiesWithFilters(ctx, filters)
	if err != nil {
		return nil, "", err
	}

	switch format {
	case "json":
		return s.exportAsJSON(logs)
	case "csv":
		return s.exportAsCSV(logs)
	default:
		return nil, "", fmt.Errorf("unsupported format: %s", format)
	}
}

// exportAsJSON converts activities to JSON format
func (s *ActivityService) exportAsJSON(logs []*models.ActivityLog) ([]byte, string, error) {
	data, err := json.MarshalIndent(logs, "", "  ")
	return data, "application/json", err
}

// exportAsCSV converts activities to CSV format
func (s *ActivityService) exportAsCSV(logs []*models.ActivityLog) ([]byte, string, error) {
	buffer := &bytes.Buffer{}
	writer := csv.NewWriter(buffer)

	// Write header
	header := []string{"Timestamp", "User Email", "Action", "Resource Type", "Resource ID", "Details"}
	if err := writer.Write(header); err != nil {
		return nil, "", err
	}

	// Write rows
	for _, log := range logs {
		details := ""
		if log.Details != nil {
			// Parse and format details
			var det map[string]interface{}
			if err := json.Unmarshal(log.Details, &det); err == nil {
				details = fmt.Sprintf("%v", det)
			}
		}

		row := []string{
			log.CreatedAt.Format("2006-01-02T15:04:05Z"),
			log.Username,
			log.Action,
			log.ResourceType,
			fmt.Sprintf("%v", log.ResourceID),
			details,
		}
		if err := writer.Write(row); err != nil {
			return nil, "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, "", err
	}

	return buffer.Bytes(), "text/csv", nil
}
