package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flexflag/flexflag/pkg/types"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type FlagRepository struct {
	db *sql.DB
}

func NewFlagRepository(db *sql.DB) *FlagRepository {
	return &FlagRepository{db: db}
}

func (r *FlagRepository) Create(ctx context.Context, flag *types.Flag) error {
	query := `
		INSERT INTO flags (
			id, project_id, key, name, description, type, enabled, 
			default_value, variations, targeting, environment, 
			tags, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)`

	// Ensure variations is never nil
	if flag.Variations == nil {
		flag.Variations = []types.Variation{}
	}
	variationsJSON, err := json.Marshal(flag.Variations)
	if err != nil {
		return fmt.Errorf("failed to marshal variations: %w", err)
	}

	var targetingJSON sql.NullString
	if flag.Targeting != nil {
		targetingBytes, err := json.Marshal(flag.Targeting)
		if err != nil {
			return fmt.Errorf("failed to marshal targeting: %w", err)
		}
		targetingJSON = sql.NullString{String: string(targetingBytes), Valid: true}
	}

	var metadataJSON sql.NullString
	if flag.Metadata != nil && len(flag.Metadata) > 0 {
		metadataBytes, err := json.Marshal(flag.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON = sql.NullString{String: string(metadataBytes), Valid: true}
	}

	now := time.Now()
	flag.CreatedAt = now
	flag.UpdatedAt = now

	_, err = r.db.ExecContext(ctx, query,
		flag.ID,
		flag.ProjectID,
		flag.Key,
		flag.Name,
		flag.Description,
		flag.Type,
		flag.Enabled,
		flag.Default,
		variationsJSON,
		targetingJSON,
		flag.Environment,
		pq.Array(flag.Tags),
		metadataJSON,
		flag.CreatedAt,
		flag.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create flag: %w", err)
	}

	return nil
}

func (r *FlagRepository) GetByKey(ctx context.Context, key, environment string) (*types.Flag, error) {
	query := `
		SELECT 
			id, project_id, key, name, description, type, enabled,
			default_value, variations, targeting, environment,
			tags, metadata, created_at, updated_at
		FROM flags
		WHERE key = $1 AND environment = $2`

	flag := &types.Flag{}
	var variationsJSON []byte
	var targetingJSON sql.NullString
	var metadataJSON sql.NullString

	err := r.db.QueryRowContext(ctx, query, key, environment).Scan(
		&flag.ID,
		&flag.ProjectID,
		&flag.Key,
		&flag.Name,
		&flag.Description,
		&flag.Type,
		&flag.Enabled,
		&flag.Default,
		&variationsJSON,
		&targetingJSON,
		&flag.Environment,
		pq.Array(&flag.Tags),
		&metadataJSON,
		&flag.CreatedAt,
		&flag.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("flag not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get flag: %w", err)
	}

	if err := json.Unmarshal(variationsJSON, &flag.Variations); err != nil {
		return nil, fmt.Errorf("failed to unmarshal variations: %w", err)
	}

	if targetingJSON.Valid {
		if err := json.Unmarshal([]byte(targetingJSON.String), &flag.Targeting); err != nil {
			return nil, fmt.Errorf("failed to unmarshal targeting: %w", err)
		}
	}

	if metadataJSON.Valid {
		if err := json.Unmarshal([]byte(metadataJSON.String), &flag.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return flag, nil
}

func (r *FlagRepository) GetByProjectKey(ctx context.Context, projectID, key, environment string) (*types.Flag, error) {
	query := `
		SELECT 
			id, project_id, key, name, description, type, enabled,
			default_value, variations, targeting, environment,
			tags, metadata, created_at, updated_at
		FROM flags
		WHERE project_id = $1 AND key = $2 AND environment = $3`

	flag := &types.Flag{}
	var variationsJSON []byte
	var targetingJSON sql.NullString
	var metadataJSON sql.NullString

	err := r.db.QueryRowContext(ctx, query, projectID, key, environment).Scan(
		&flag.ID,
		&flag.ProjectID,
		&flag.Key,
		&flag.Name,
		&flag.Description,
		&flag.Type,
		&flag.Enabled,
		&flag.Default,
		&variationsJSON,
		&targetingJSON,
		&flag.Environment,
		pq.Array(&flag.Tags),
		&metadataJSON,
		&flag.CreatedAt,
		&flag.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("flag not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get flag: %w", err)
	}

	if err := json.Unmarshal(variationsJSON, &flag.Variations); err != nil {
		return nil, fmt.Errorf("failed to unmarshal variations: %w", err)
	}

	if targetingJSON.Valid {
		if err := json.Unmarshal([]byte(targetingJSON.String), &flag.Targeting); err != nil {
			return nil, fmt.Errorf("failed to unmarshal targeting: %w", err)
		}
	}

	if metadataJSON.Valid {
		if err := json.Unmarshal([]byte(metadataJSON.String), &flag.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return flag, nil
}

func (r *FlagRepository) List(ctx context.Context, environment string) ([]*types.Flag, error) {
	query := `
		SELECT 
			id, project_id, key, name, description, type, enabled,
			default_value, variations, targeting, environment,
			tags, metadata, created_at, updated_at
		FROM flags
		WHERE environment = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, environment)
	if err != nil {
		return nil, fmt.Errorf("failed to list flags: %w", err)
	}
	defer rows.Close()

	var flags []*types.Flag
	for rows.Next() {
		flag := &types.Flag{}
		var variationsJSON []byte
		var targetingJSON sql.NullString
		var metadataJSON sql.NullString

		err := rows.Scan(
			&flag.ID,
			&flag.ProjectID,
			&flag.Key,
			&flag.Name,
			&flag.Description,
			&flag.Type,
			&flag.Enabled,
			&flag.Default,
			&variationsJSON,
			&targetingJSON,
			&flag.Environment,
			pq.Array(&flag.Tags),
			&metadataJSON,
			&flag.CreatedAt,
			&flag.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan flag: %w", err)
		}

		if err := json.Unmarshal(variationsJSON, &flag.Variations); err != nil {
			return nil, fmt.Errorf("failed to unmarshal variations: %w", err)
		}

		if targetingJSON.Valid {
			if err := json.Unmarshal([]byte(targetingJSON.String), &flag.Targeting); err != nil {
				return nil, fmt.Errorf("failed to unmarshal targeting: %w", err)
			}
		}

		if metadataJSON.Valid {
			if err := json.Unmarshal([]byte(metadataJSON.String), &flag.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		flags = append(flags, flag)
	}

	return flags, nil
}

// ListByProject retrieves flags filtered by project_id and environment
func (r *FlagRepository) ListByProject(ctx context.Context, projectID, environment string) ([]*types.Flag, error) {
	query := `
		SELECT 
			id, project_id, key, name, description, type, enabled,
			default_value, variations, targeting, environment,
			tags, metadata, created_at, updated_at
		FROM flags
		WHERE project_id = $1 AND environment = $2
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, projectID, environment)
	if err != nil {
		return nil, fmt.Errorf("failed to list flags by project: %w", err)
	}
	defer rows.Close()

	var flags []*types.Flag
	for rows.Next() {
		flag := &types.Flag{}
		var variationsJSON []byte
		var targetingJSON sql.NullString
		var metadataJSON sql.NullString

		err := rows.Scan(
			&flag.ID,
			&flag.ProjectID,
			&flag.Key,
			&flag.Name,
			&flag.Description,
			&flag.Type,
			&flag.Enabled,
			&flag.Default,
			&variationsJSON,
			&targetingJSON,
			&flag.Environment,
			pq.Array(&flag.Tags),
			&metadataJSON,
			&flag.CreatedAt,
			&flag.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan flag: %w", err)
		}

		if err := json.Unmarshal(variationsJSON, &flag.Variations); err != nil {
			return nil, fmt.Errorf("failed to unmarshal variations: %w", err)
		}

		if targetingJSON.Valid {
			if err := json.Unmarshal([]byte(targetingJSON.String), &flag.Targeting); err != nil {
				return nil, fmt.Errorf("failed to unmarshal targeting: %w", err)
			}
		}

		if metadataJSON.Valid {
			if err := json.Unmarshal([]byte(metadataJSON.String), &flag.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		} else {
			flag.Metadata = make(map[string]interface{})
		}

		flags = append(flags, flag)
	}

	return flags, nil
}

func (r *FlagRepository) Update(ctx context.Context, flag *types.Flag) error {
	query := `
		UPDATE flags SET
			name = $3,
			description = $4,
			type = $5,
			enabled = $6,
			default_value = $7,
			variations = $8,
			targeting = $9,
			tags = $10,
			metadata = $11,
			updated_at = $12
		WHERE key = $1 AND environment = $2`

	// Ensure variations is never nil
	if flag.Variations == nil {
		flag.Variations = []types.Variation{}
	}
	variationsJSON, err := json.Marshal(flag.Variations)
	if err != nil {
		return fmt.Errorf("failed to marshal variations: %w", err)
	}

	var targetingJSON sql.NullString
	if flag.Targeting != nil {
		targetingBytes, err := json.Marshal(flag.Targeting)
		if err != nil {
			return fmt.Errorf("failed to marshal targeting: %w", err)
		}
		targetingJSON = sql.NullString{String: string(targetingBytes), Valid: true}
	}

	var metadataJSON sql.NullString
	if flag.Metadata != nil && len(flag.Metadata) > 0 {
		metadataBytes, err := json.Marshal(flag.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON = sql.NullString{String: string(metadataBytes), Valid: true}
	}

	flag.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		flag.Key,
		flag.Environment,
		flag.Name,
		flag.Description,
		flag.Type,
		flag.Enabled,
		flag.Default,
		variationsJSON,
		targetingJSON,
		pq.Array(flag.Tags),
		metadataJSON,
		flag.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update flag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("flag not found")
	}

	return nil
}

func (r *FlagRepository) Delete(ctx context.Context, key, environment string) error {
	query := `DELETE FROM flags WHERE key = $1 AND environment = $2`

	result, err := r.db.ExecContext(ctx, query, key, environment)
	if err != nil {
		return fmt.Errorf("failed to delete flag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("flag not found")
	}

	return nil
}

func (r *FlagRepository) ListByTags(ctx context.Context, environment string, tags []string) ([]*types.Flag, error) {
	query := `
		SELECT 
			id, key, name, description, type, enabled,
			default_value, variations, targeting, environment,
			tags, metadata, created_at, updated_at
		FROM flags
		WHERE environment = $1 AND tags && $2
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, environment, pq.Array(tags))
	if err != nil {
		return nil, fmt.Errorf("failed to list flags by tags: %w", err)
	}
	defer rows.Close()

	var flags []*types.Flag
	for rows.Next() {
		flag := &types.Flag{}
		var variationsJSON []byte
		var targetingJSON sql.NullString
		var metadataJSON sql.NullString

		err := rows.Scan(
			&flag.ID,
			&flag.ProjectID,
			&flag.Key,
			&flag.Name,
			&flag.Description,
			&flag.Type,
			&flag.Enabled,
			&flag.Default,
			&variationsJSON,
			&targetingJSON,
			&flag.Environment,
			pq.Array(&flag.Tags),
			&metadataJSON,
			&flag.CreatedAt,
			&flag.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan flag: %w", err)
		}

		if err := json.Unmarshal(variationsJSON, &flag.Variations); err != nil {
			return nil, fmt.Errorf("failed to unmarshal variations: %w", err)
		}

		if targetingJSON.Valid {
			if err := json.Unmarshal([]byte(targetingJSON.String), &flag.Targeting); err != nil {
				return nil, fmt.Errorf("failed to unmarshal targeting: %w", err)
			}
		}

		if metadataJSON.Valid {
			if err := json.Unmarshal([]byte(metadataJSON.String), &flag.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		flags = append(flags, flag)
	}

	return flags, nil
}