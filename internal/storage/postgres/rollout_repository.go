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

type RolloutRepository struct {
	db *sql.DB
}

func NewRolloutRepository(db *sql.DB) *RolloutRepository {
	return &RolloutRepository{db: db}
}

// Create creates a new rollout
func (r *RolloutRepository) Create(ctx context.Context, rollout *types.Rollout) error {
	rollout.ID = uuid.New().String()
	
	query := `
		INSERT INTO rollouts (id, flag_id, environment, type, name, description, config, status, start_date, end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
	`
	
	configJSON, err := json.Marshal(rollout.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	now := time.Now()
	rollout.CreatedAt = now
	rollout.UpdatedAt = now
	
	_, err = r.db.ExecContext(ctx, query,
		rollout.ID, rollout.FlagID, rollout.Environment, rollout.Type,
		rollout.Name, rollout.Description, configJSON, rollout.Status,
		rollout.StartDate, rollout.EndDate)
	if err != nil {
		return fmt.Errorf("failed to create rollout: %w", err)
	}
	
	return nil
}

// GetByID retrieves a rollout by ID
func (r *RolloutRepository) GetByID(ctx context.Context, id string) (*types.Rollout, error) {
	query := `
		SELECT id, flag_id, environment, type, name, description, config, status, start_date, end_date, created_at, updated_at
		FROM rollouts
		WHERE id = $1
	`
	
	rollout := &types.Rollout{}
	var configJSON []byte
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rollout.ID, &rollout.FlagID, &rollout.Environment, &rollout.Type,
		&rollout.Name, &rollout.Description, &configJSON, &rollout.Status,
		&rollout.StartDate, &rollout.EndDate, &rollout.CreatedAt, &rollout.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("rollout not found")
		}
		return nil, fmt.Errorf("failed to get rollout: %w", err)
	}
	
	if err := json.Unmarshal(configJSON, &rollout.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	return rollout, nil
}

// GetByFlag retrieves all rollouts for a specific flag and environment
func (r *RolloutRepository) GetByFlag(ctx context.Context, flagID, environment string) ([]*types.Rollout, error) {
	query := `
		SELECT id, flag_id, environment, type, name, description, config, status, start_date, end_date, created_at, updated_at
		FROM rollouts
		WHERE flag_id = $1 AND environment = $2
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query, flagID, environment)
	if err != nil {
		return nil, fmt.Errorf("failed to get rollouts: %w", err)
	}
	defer rows.Close()
	
	var rollouts []*types.Rollout
	for rows.Next() {
		rollout := &types.Rollout{}
		var configJSON []byte
		
		err := rows.Scan(
			&rollout.ID, &rollout.FlagID, &rollout.Environment, &rollout.Type,
			&rollout.Name, &rollout.Description, &configJSON, &rollout.Status,
			&rollout.StartDate, &rollout.EndDate, &rollout.CreatedAt, &rollout.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rollout: %w", err)
		}
		
		if err := json.Unmarshal(configJSON, &rollout.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
		
		rollouts = append(rollouts, rollout)
	}
	
	return rollouts, nil
}

// GetByProject retrieves all rollouts for a specific project and environment
func (r *RolloutRepository) GetByProject(ctx context.Context, projectID, environment string) ([]*types.Rollout, error) {
	query := `
		SELECT r.id, r.flag_id, r.environment, r.type, r.name, r.description, r.config, r.status, r.start_date, r.end_date, r.created_at, r.updated_at
		FROM rollouts r
		JOIN flags f ON r.flag_id = f.id
		WHERE f.project_id = $1 AND r.environment = $2
		ORDER BY r.created_at DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query, projectID, environment)
	if err != nil {
		return nil, fmt.Errorf("failed to get rollouts: %w", err)
	}
	defer rows.Close()
	
	var rollouts []*types.Rollout
	for rows.Next() {
		rollout := &types.Rollout{}
		var configJSON []byte
		
		err := rows.Scan(
			&rollout.ID, &rollout.FlagID, &rollout.Environment, &rollout.Type,
			&rollout.Name, &rollout.Description, &configJSON, &rollout.Status,
			&rollout.StartDate, &rollout.EndDate, &rollout.CreatedAt, &rollout.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rollout: %w", err)
		}
		
		if err := json.Unmarshal(configJSON, &rollout.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
		
		rollouts = append(rollouts, rollout)
	}
	
	return rollouts, nil
}

// GetActiveRollouts retrieves all active rollouts for a flag and environment
func (r *RolloutRepository) GetActiveRollouts(ctx context.Context, flagID, environment string) ([]*types.Rollout, error) {
	query := `
		SELECT id, flag_id, environment, type, name, description, config, status, start_date, end_date, created_at, updated_at
		FROM rollouts
		WHERE flag_id = $1 AND environment = $2 AND status = 'active'
		AND (start_date IS NULL OR start_date <= NOW())
		AND (end_date IS NULL OR end_date > NOW())
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query, flagID, environment)
	if err != nil {
		return nil, fmt.Errorf("failed to get active rollouts: %w", err)
	}
	defer rows.Close()
	
	var rollouts []*types.Rollout
	for rows.Next() {
		rollout := &types.Rollout{}
		var configJSON []byte
		
		err := rows.Scan(
			&rollout.ID, &rollout.FlagID, &rollout.Environment, &rollout.Type,
			&rollout.Name, &rollout.Description, &configJSON, &rollout.Status,
			&rollout.StartDate, &rollout.EndDate, &rollout.CreatedAt, &rollout.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rollout: %w", err)
		}
		
		if err := json.Unmarshal(configJSON, &rollout.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
		
		rollouts = append(rollouts, rollout)
	}
	
	return rollouts, nil
}

// Update updates a rollout
func (r *RolloutRepository) Update(ctx context.Context, rollout *types.Rollout) error {
	query := `
		UPDATE rollouts SET
			name = $2,
			description = $3,
			config = $4,
			status = $5,
			start_date = $6,
			end_date = $7,
			updated_at = NOW()
		WHERE id = $1
	`
	
	configJSON, err := json.Marshal(rollout.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	rollout.UpdatedAt = time.Now()
	
	result, err := r.db.ExecContext(ctx, query,
		rollout.ID, rollout.Name, rollout.Description, configJSON,
		rollout.Status, rollout.StartDate, rollout.EndDate)
	if err != nil {
		return fmt.Errorf("failed to update rollout: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("rollout not found")
	}
	
	return nil
}

// Delete deletes a rollout
func (r *RolloutRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM rollouts WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete rollout: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("rollout not found")
	}
	
	return nil
}

// Sticky Assignment methods

// CreateStickyAssignment creates a new sticky assignment
func (r *RolloutRepository) CreateStickyAssignment(ctx context.Context, assignment *types.StickyAssignment) error {
	assignment.ID = uuid.New().String()
	
	query := `
		INSERT INTO sticky_assignments (id, flag_id, environment, user_id, user_key, variation_id, bucket_key, assigned_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), $8)
		ON CONFLICT (flag_id, environment, user_key)
		DO UPDATE SET
			variation_id = EXCLUDED.variation_id,
			bucket_key = EXCLUDED.bucket_key,
			assigned_at = NOW(),
			expires_at = EXCLUDED.expires_at
	`
	
	assignment.AssignedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		assignment.ID, assignment.FlagID, assignment.Environment,
		assignment.UserID, assignment.UserKey, assignment.VariationID,
		assignment.BucketKey, assignment.ExpiresAt)
	if err != nil {
		return fmt.Errorf("failed to create sticky assignment: %w", err)
	}
	
	return nil
}

// GetStickyAssignment retrieves a sticky assignment for a user and flag
func (r *RolloutRepository) GetStickyAssignment(ctx context.Context, flagID, environment, userKey string) (*types.StickyAssignment, error) {
	query := `
		SELECT id, flag_id, environment, user_id, user_key, variation_id, bucket_key, assigned_at, expires_at
		FROM sticky_assignments
		WHERE flag_id = $1 AND environment = $2 AND user_key = $3
		AND (expires_at IS NULL OR expires_at > NOW())
	`
	
	assignment := &types.StickyAssignment{}
	
	err := r.db.QueryRowContext(ctx, query, flagID, environment, userKey).Scan(
		&assignment.ID, &assignment.FlagID, &assignment.Environment,
		&assignment.UserID, &assignment.UserKey, &assignment.VariationID,
		&assignment.BucketKey, &assignment.AssignedAt, &assignment.ExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No assignment found is not an error
		}
		return nil, fmt.Errorf("failed to get sticky assignment: %w", err)
	}
	
	return assignment, nil
}

// DeleteStickyAssignment removes a sticky assignment
func (r *RolloutRepository) DeleteStickyAssignment(ctx context.Context, flagID, environment, userKey string) error {
	query := `DELETE FROM sticky_assignments WHERE flag_id = $1 AND environment = $2 AND user_key = $3`
	
	_, err := r.db.ExecContext(ctx, query, flagID, environment, userKey)
	if err != nil {
		return fmt.Errorf("failed to delete sticky assignment: %w", err)
	}
	
	return nil
}

// CleanupExpiredAssignments removes expired sticky assignments
func (r *RolloutRepository) CleanupExpiredAssignments(ctx context.Context) error {
	query := `DELETE FROM sticky_assignments WHERE expires_at IS NOT NULL AND expires_at <= NOW()`
	
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired assignments: %w", err)
	}
	
	return nil
}