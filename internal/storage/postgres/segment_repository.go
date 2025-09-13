package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flexflag/flexflag/pkg/types"
	"github.com/google/uuid"
)

type SegmentRepository struct {
	db *sql.DB
}

func NewSegmentRepository(db *sql.DB) *SegmentRepository {
	return &SegmentRepository{db: db}
}

// Create creates a new segment
func (r *SegmentRepository) Create(ctx context.Context, segment *types.Segment) error {
	segment.ID = uuid.New().String()
	
	query := `
		INSERT INTO segments (id, project_id, key, name, description, rules, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`
	
	rulesJSON, err := json.Marshal(segment.Rules)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}
	
	now := time.Now()
	segment.CreatedAt = now
	segment.UpdatedAt = now
	
	_, err = r.db.ExecContext(ctx, query,
		segment.ID, segment.ProjectID, segment.Key, segment.Name, segment.Description,
		rulesJSON)
	if err != nil {
		return fmt.Errorf("failed to create segment: %w", err)
	}
	
	return nil
}

// GetByKey retrieves a segment by project_id and key
func (r *SegmentRepository) GetByKey(ctx context.Context, projectID, key string) (*types.Segment, error) {
	query := `
		SELECT id, project_id, key, name, description, rules, created_at, updated_at
		FROM segments
		WHERE project_id = $1 AND key = $2
	`
	
	segment := &types.Segment{}
	var rulesJSON []byte
	
	err := r.db.QueryRowContext(ctx, query, projectID, key).Scan(
		&segment.ID, &segment.ProjectID, &segment.Key, &segment.Name, &segment.Description,
		&rulesJSON, &segment.CreatedAt, &segment.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("segment not found")
		}
		return nil, fmt.Errorf("failed to get segment: %w", err)
	}
	
	if err := json.Unmarshal(rulesJSON, &segment.Rules); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
	}
	
	return segment, nil
}

// GetByID retrieves a segment by ID
func (r *SegmentRepository) GetByID(ctx context.Context, id string) (*types.Segment, error) {
	query := `
		SELECT id, project_id, key, name, description, rules, created_at, updated_at
		FROM segments
		WHERE id = $1
	`
	
	segment := &types.Segment{}
	var rulesJSON []byte
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&segment.ID, &segment.ProjectID, &segment.Key, &segment.Name, &segment.Description,
		&rulesJSON, &segment.CreatedAt, &segment.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("segment not found")
		}
		return nil, fmt.Errorf("failed to get segment: %w", err)
	}
	
	if err := json.Unmarshal(rulesJSON, &segment.Rules); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
	}
	
	return segment, nil
}

// List retrieves all segments for a project with pagination
func (r *SegmentRepository) List(ctx context.Context, projectID string, limit, offset int) ([]*types.Segment, error) {
	query := `
		SELECT id, project_id, key, name, description, rules, created_at, updated_at
		FROM segments
		WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.QueryContext(ctx, query, projectID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list segments: %w", err)
	}
	defer rows.Close()
	
	var segments []*types.Segment
	for rows.Next() {
		segment := &types.Segment{}
		var rulesJSON []byte
		
		err := rows.Scan(
			&segment.ID, &segment.ProjectID, &segment.Key, &segment.Name, &segment.Description,
			&rulesJSON, &segment.CreatedAt, &segment.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan segment: %w", err)
		}
		
		if err := json.Unmarshal(rulesJSON, &segment.Rules); err != nil {
			return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
		}
		
		segments = append(segments, segment)
	}
	
	return segments, nil
}

// Update updates a segment
func (r *SegmentRepository) Update(ctx context.Context, segment *types.Segment) error {
	query := `
		UPDATE segments SET
			name = $3,
			description = $4,
			rules = $5,
			updated_at = NOW()
		WHERE project_id = $1 AND key = $2
	`
	
	rulesJSON, err := json.Marshal(segment.Rules)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}
	
	segment.UpdatedAt = time.Now()
	
	result, err := r.db.ExecContext(ctx, query,
		segment.ProjectID, segment.Key, segment.Name, segment.Description, rulesJSON)
	if err != nil {
		return fmt.Errorf("failed to update segment: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("segment not found")
	}
	
	return nil
}

// Delete deletes a segment
func (r *SegmentRepository) Delete(ctx context.Context, projectID, key string) error {
	query := `DELETE FROM segments WHERE project_id = $1 AND key = $2`
	
	result, err := r.db.ExecContext(ctx, query, projectID, key)
	if err != nil {
		return fmt.Errorf("failed to delete segment: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("segment not found")
	}
	
	return nil
}

// KeyExists checks if a segment key already exists in a project
func (r *SegmentRepository) KeyExists(ctx context.Context, projectID, key string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM segments WHERE project_id = $1 AND key = $2`
	
	err := r.db.QueryRowContext(ctx, query, projectID, key).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}
	
	return count > 0, nil
}

// GetByKeys retrieves multiple segments by their keys in a single project
func (r *SegmentRepository) GetByKeys(ctx context.Context, keys []string, projectID string) ([]*types.Segment, error) {
	if len(keys) == 0 {
		return []*types.Segment{}, nil
	}
	
	query := `
		SELECT id, project_id, key, name, description, rules, created_at, updated_at
		FROM segments
		WHERE project_id = $1 AND key = ANY($2)
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query, projectID, keys) // PostgreSQL-specific array syntax
	if err != nil {
		return nil, fmt.Errorf("failed to get segments by keys: %w", err)
	}
	defer rows.Close()
	
	var segments []*types.Segment
	for rows.Next() {
		segment := &types.Segment{}
		var rulesJSON []byte
		
		err := rows.Scan(
			&segment.ID, &segment.ProjectID, &segment.Key, &segment.Name, &segment.Description,
			&rulesJSON, &segment.CreatedAt, &segment.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan segment: %w", err)
		}
		
		if err := json.Unmarshal(rulesJSON, &segment.Rules); err != nil {
			return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
		}
		
		segments = append(segments, segment)
	}
	
	return segments, nil
}