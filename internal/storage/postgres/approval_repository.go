package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flexflag/flexflag/pkg/types"
)

type ApprovalRepository struct {
	db *sql.DB
}

func NewApprovalRepository(db *sql.DB) *ApprovalRepository {
	return &ApprovalRepository{db: db}
}

func (r *ApprovalRepository) Create(ctx context.Context, req *types.ApprovalRequest) error {
	payload, err := json.Marshal(req.Payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	query := `
		INSERT INTO flag_approval_requests
			(id, project_id, flag_key, environment, change_type, payload,
			 requested_by, requested_by_name, status, created_at, updated_at)
		VALUES
			(gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, 'pending', NOW(), NOW())
		RETURNING id, created_at, updated_at`

	row := r.db.QueryRowContext(ctx, query,
		req.ProjectID, req.FlagKey, req.Environment, req.ChangeType,
		payload, nullString(req.RequestedBy), req.RequestedByName,
	)
	return row.Scan(&req.ID, &req.CreatedAt, &req.UpdatedAt)
}

func (r *ApprovalRepository) GetByID(ctx context.Context, id string) (*types.ApprovalRequest, error) {
	query := `
		SELECT id, project_id, flag_key, environment, change_type, payload,
		       requested_by, requested_by_name, status, review_note,
		       reviewed_by, reviewed_at, created_at, updated_at
		FROM flag_approval_requests
		WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)
	return scanApproval(row)
}

func (r *ApprovalRepository) ListByProject(ctx context.Context, projectID string, status types.ApprovalStatus) ([]*types.ApprovalRequest, error) {
	var (
		rows *sql.Rows
		err  error
	)

	base := `
		SELECT id, project_id, flag_key, environment, change_type, payload,
		       requested_by, requested_by_name, status, review_note,
		       reviewed_by, reviewed_at, created_at, updated_at
		FROM flag_approval_requests
		WHERE project_id = $1`

	if status != "" {
		rows, err = r.db.QueryContext(ctx, base+" AND status = $2 ORDER BY created_at DESC", projectID, string(status))
	} else {
		rows, err = r.db.QueryContext(ctx, base+" ORDER BY created_at DESC", projectID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*types.ApprovalRequest
	for rows.Next() {
		req, err := scanApprovalRow(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, req)
	}
	return results, rows.Err()
}

func (r *ApprovalRepository) UpdateStatus(ctx context.Context, id string, status types.ApprovalStatus, reviewedBy, note string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `
		UPDATE flag_approval_requests
		SET status = $1, review_note = $2, reviewed_by = $3, reviewed_at = $4, updated_at = $4
		WHERE id = $5`,
		string(status), note, nullString(reviewedBy), now, id,
	)
	return err
}

func (r *ApprovalRepository) PendingCount(ctx context.Context, projectID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM flag_approval_requests WHERE project_id = $1 AND status = 'pending'`,
		projectID,
	).Scan(&count)
	return count, err
}

// ── helpers ──────────────────────────────────────────────────────────────────

func scanApproval(row *sql.Row) (*types.ApprovalRequest, error) {
	var (
		req         types.ApprovalRequest
		payloadJSON []byte
		requestedBy sql.NullString
		reviewedBy  sql.NullString
		reviewNote  sql.NullString
		reviewedAt  sql.NullTime
	)
	err := row.Scan(
		&req.ID, &req.ProjectID, &req.FlagKey, &req.Environment,
		&req.ChangeType, &payloadJSON,
		&requestedBy, &req.RequestedByName, &req.Status,
		&reviewNote, &reviewedBy, &reviewedAt,
		&req.CreatedAt, &req.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(payloadJSON, &req.Payload)
	req.RequestedBy = requestedBy.String
	req.ReviewedBy = reviewedBy.String
	req.ReviewNote = reviewNote.String
	if reviewedAt.Valid {
		t := reviewedAt.Time
		req.ReviewedAt = &t
	}
	return &req, nil
}

func scanApprovalRow(rows *sql.Rows) (*types.ApprovalRequest, error) {
	var (
		req         types.ApprovalRequest
		payloadJSON []byte
		requestedBy sql.NullString
		reviewedBy  sql.NullString
		reviewNote  sql.NullString
		reviewedAt  sql.NullTime
	)
	err := rows.Scan(
		&req.ID, &req.ProjectID, &req.FlagKey, &req.Environment,
		&req.ChangeType, &payloadJSON,
		&requestedBy, &req.RequestedByName, &req.Status,
		&reviewNote, &reviewedBy, &reviewedAt,
		&req.CreatedAt, &req.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(payloadJSON, &req.Payload)
	req.RequestedBy = requestedBy.String
	req.ReviewedBy = reviewedBy.String
	req.ReviewNote = reviewNote.String
	if reviewedAt.Valid {
		t := reviewedAt.Time
		req.ReviewedAt = &t
	}
	return &req, nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
