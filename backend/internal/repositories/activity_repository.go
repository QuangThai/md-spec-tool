package repositories

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/md-spec-tool/internal/models"
)

type ActivityRepository struct {
	pool *pgxpool.Pool
}

func NewActivityRepository(pool *pgxpool.Pool) *ActivityRepository {
	return &ActivityRepository{pool: pool}
}

// Log creates a new activity log entry
func (r *ActivityRepository) Log(ctx context.Context, specID, userID, action, resourceType string, resourceID *string, details *models.ActivityDetails) error {
	query := `
		INSERT INTO activity_logs (spec_id, user_id, action, resource_type, resource_id, details)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	detailsJSON, _ := json.Marshal(details)

	_, err := r.pool.Exec(ctx, query, specID, userID, action, resourceType, resourceID, detailsJSON)
	return err
}

// GetBySpecID retrieves activity logs for a spec
func (r *ActivityRepository) GetBySpecID(ctx context.Context, specID string, limit int, offset int) ([]*models.ActivityLog, int, error) {
	// Get total count
	var count int
	countQuery := `SELECT COUNT(*) FROM activity_logs WHERE spec_id = $1`
	err := r.pool.QueryRow(ctx, countQuery, specID).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated logs with username
	query := `
		SELECT a.id, a.spec_id, a.user_id, u.username, a.action, a.resource_type, a.resource_id, a.details, a.created_at
		FROM activity_logs a
		JOIN users u ON a.user_id = u.id
		WHERE a.spec_id = $1
		ORDER BY a.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, specID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*models.ActivityLog
	for rows.Next() {
		log := &models.ActivityLog{}
		err := rows.Scan(&log.ID, &log.SpecID, &log.UserID, &log.Username,
			&log.Action, &log.ResourceType, &log.ResourceID, &log.Details, &log.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}

	return logs, count, rows.Err()
}

// GetByUserID retrieves activity logs for a user (specs they own/collaborate on)
func (r *ActivityRepository) GetByUserID(ctx context.Context, userID string, limit int, offset int) ([]*models.ActivityLog, int, error) {
	// Get count
	var count int
	countQuery := `
		SELECT COUNT(*) FROM activity_logs
		WHERE user_id = $1 OR spec_id IN (
			SELECT id FROM specs WHERE user_id = $1
		)
	`
	err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	// Get logs
	query := `
		SELECT a.id, a.spec_id, a.user_id, u.username, a.action, a.resource_type, a.resource_id, a.details, a.created_at
		FROM activity_logs a
		JOIN users u ON a.user_id = u.id
		WHERE a.user_id = $1 OR a.spec_id IN (
			SELECT id FROM specs WHERE user_id = $1
		)
		ORDER BY a.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*models.ActivityLog
	for rows.Next() {
		log := &models.ActivityLog{}
		err := rows.Scan(&log.ID, &log.SpecID, &log.UserID, &log.Username,
			&log.Action, &log.ResourceType, &log.ResourceID, &log.Details, &log.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}

	return logs, count, rows.Err()
}

// PHASE 7: Activity filtering and statistics

// GetActivitiesWithFilters retrieves activity logs with dynamic filtering
func (r *ActivityRepository) GetActivitiesWithFilters(ctx context.Context, filters *models.ActivityFilterRequest) ([]*models.ActivityLog, int, error) {
	// Build dynamic WHERE clause
	var whereConditions []string
	var args []interface{}
	argCount := 1

	// Required: spec_id
	whereConditions = append(whereConditions, fmt.Sprintf("a.spec_id = $%d", argCount))
	args = append(args, filters.SpecID)
	argCount++

	// Optional: actions (OR logic within group)
	if len(filters.Actions) > 0 {
		actionCondition := "a.action IN ("
		for i, action := range filters.Actions {
			if i > 0 {
				actionCondition += ", "
			}
			actionCondition += fmt.Sprintf("$%d", argCount)
			args = append(args, action)
			argCount++
		}
		actionCondition += ")"
		whereConditions = append(whereConditions, actionCondition)
	}

	// Optional: user_id filter
	if filters.UserID != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("a.user_id = $%d", argCount))
		args = append(args, filters.UserID)
		argCount++
	}

	// Optional: resource_type filter
	if filters.ResourceType != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("a.resource_type = $%d", argCount))
		args = append(args, filters.ResourceType)
		argCount++
	}

	// Optional: date range
	if filters.StartDate != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("a.created_at >= $%d", argCount))
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("a.created_at <= $%d", argCount))
		args = append(args, filters.EndDate)
		argCount++
	}

	// Build WHERE clause
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + whereConditions[0]
		for _, cond := range whereConditions[1:] {
			whereClause += " AND " + cond
		}
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM activity_logs a " + whereClause
	var count int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated logs
	limit := filters.Limit
	if limit == 0 {
		limit = 50
	}
	offset := filters.Offset

	query := fmt.Sprintf(`
		SELECT a.id, a.spec_id, a.user_id, u.username, a.action, a.resource_type, a.resource_id, a.details, a.filters_applied, a.created_at
		FROM activity_logs a
		JOIN users u ON a.user_id = u.id
		%s
		ORDER BY a.created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, argCount, argCount+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*models.ActivityLog
	for rows.Next() {
		log := &models.ActivityLog{}
		err := rows.Scan(&log.ID, &log.SpecID, &log.UserID, &log.Username,
			&log.Action, &log.ResourceType, &log.ResourceID, &log.Details, &log.FiltersApplied, &log.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}

	return logs, count, rows.Err()
}

// GetActivityStats returns count of activities by action type for a spec
func (r *ActivityRepository) GetActivityStats(ctx context.Context, specID string) (*models.ActivityStatsResponse, error) {
	query := `
		SELECT
			COUNT(CASE WHEN action = 'created' THEN 1 END) as created,
			COUNT(CASE WHEN action = 'updated' THEN 1 END) as updated,
			COUNT(CASE WHEN action = 'shared' THEN 1 END) as shared,
			COUNT(CASE WHEN action = 'commented' THEN 1 END) as commented,
			COUNT(CASE WHEN action = 'deleted' THEN 1 END) as deleted
		FROM activity_logs
		WHERE spec_id = $1
	`

	stats := &models.ActivityStatsResponse{}
	err := r.pool.QueryRow(ctx, query, specID).Scan(
		&stats.Created, &stats.Updated, &stats.Shared, &stats.Commented, &stats.Deleted,
	)

	return stats, err
}
