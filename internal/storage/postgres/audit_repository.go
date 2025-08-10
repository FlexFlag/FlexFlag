package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/flexflag/flexflag/pkg/types"
)

type AuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{
		db: db,
	}
}

func (r *AuditRepository) Create(ctx context.Context, log *types.AuditLog) error {
	query := `
		INSERT INTO audit_logs (
			project_id, user_id, resource_type, resource_id, action,
			old_values, new_values, metadata, ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	var oldValuesJSON, newValuesJSON, metadataJSON []byte
	var err error

	if log.OldValues != nil {
		oldValuesJSON = log.OldValues
	}
	
	if log.NewValues != nil {
		newValuesJSON = log.NewValues
	}

	if log.Metadata != nil {
		metadataJSON, err = json.Marshal(log.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	// Convert JSON bytes to strings for PostgreSQL JSONB columns
	var oldValuesStr, newValuesStr, metadataStr interface{}
	
	if oldValuesJSON != nil {
		oldValuesStr = string(oldValuesJSON)
	}
	
	if newValuesJSON != nil {
		newValuesStr = string(newValuesJSON)
	}
	
	if metadataJSON != nil {
		metadataStr = string(metadataJSON)
	}

	_, err = r.db.ExecContext(ctx, query,
		log.ProjectID, log.UserID, log.ResourceType, log.ResourceID, log.Action,
		oldValuesStr, newValuesStr, metadataStr, log.IPAddress, log.UserAgent,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

func (r *AuditRepository) List(ctx context.Context, projectID string, limit, offset int) ([]*types.AuditLog, error) {
	query := `
		SELECT 
			id, project_id, user_id, resource_type, resource_id, action,
			old_values, new_values, metadata, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, projectID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*types.AuditLog
	for rows.Next() {
		log := &types.AuditLog{}
		var oldValues, newValues, metadata sql.NullString

		err := rows.Scan(
			&log.ID,
			&log.ProjectID,
			&log.UserID,
			&log.ResourceType,
			&log.ResourceID,
			&log.Action,
			&oldValues,
			&newValues,
			&metadata,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if oldValues.Valid {
			log.OldValues = json.RawMessage(oldValues.String)
		}
		if newValues.Valid {
			log.NewValues = json.RawMessage(newValues.String)
		}
		if metadata.Valid {
			if err := json.Unmarshal([]byte(metadata.String), &log.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		logs = append(logs, log)
	}

	return logs, nil
}

func (r *AuditRepository) ListByResource(ctx context.Context, resourceType, resourceID string, limit, offset int) ([]*types.AuditLog, error) {
	query := `
		SELECT 
			id, project_id, user_id, resource_type, resource_id, action,
			old_values, new_values, metadata, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE resource_type = $1 AND resource_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.QueryContext(ctx, query, resourceType, resourceID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs by resource: %w", err)
	}
	defer rows.Close()

	var logs []*types.AuditLog
	for rows.Next() {
		log := &types.AuditLog{}
		var oldValues, newValues, metadata sql.NullString

		err := rows.Scan(
			&log.ID,
			&log.ProjectID,
			&log.UserID,
			&log.ResourceType,
			&log.ResourceID,
			&log.Action,
			&oldValues,
			&newValues,
			&metadata,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if oldValues.Valid {
			log.OldValues = json.RawMessage(oldValues.String)
		}
		if newValues.Valid {
			log.NewValues = json.RawMessage(newValues.String)
		}
		if metadata.Valid {
			if err := json.Unmarshal([]byte(metadata.String), &log.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		logs = append(logs, log)
	}

	return logs, nil
}

func (r *AuditRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*types.AuditLog, error) {
	query := `
		SELECT 
			id, project_id, user_id, resource_type, resource_id, action,
			old_values, new_values, metadata, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs by user: %w", err)
	}
	defer rows.Close()

	var logs []*types.AuditLog
	for rows.Next() {
		log := &types.AuditLog{}
		var oldValues, newValues, metadata sql.NullString

		err := rows.Scan(
			&log.ID,
			&log.ProjectID,
			&log.UserID,
			&log.ResourceType,
			&log.ResourceID,
			&log.Action,
			&oldValues,
			&newValues,
			&metadata,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if oldValues.Valid {
			log.OldValues = json.RawMessage(oldValues.String)
		}
		if newValues.Valid {
			log.NewValues = json.RawMessage(newValues.String)
		}
		if metadata.Valid {
			if err := json.Unmarshal([]byte(metadata.String), &log.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		logs = append(logs, log)
	}

	return logs, nil
}