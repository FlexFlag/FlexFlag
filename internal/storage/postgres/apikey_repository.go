package postgres

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/flexflag/flexflag/pkg/types"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ApiKeyRepository struct {
	db *sql.DB
}

func NewApiKeyRepository(db *sql.DB) *ApiKeyRepository {
	return &ApiKeyRepository{db: db}
}

// GenerateApiKey creates a new API key with format: ff_{env_key}_{random}
func (r *ApiKeyRepository) GenerateApiKey(ctx context.Context, req *types.CreateApiKeyRequest, projectID, userID string) (*types.ApiKey, error) {
	// Get environment details
	env, err := r.getEnvironmentByID(ctx, req.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment: %w", err)
	}

	// Generate random key
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}
	randomString := hex.EncodeToString(randomBytes)

	// Create the full API key
	fullKey := fmt.Sprintf("ff_%s_%s", env.Key, randomString)
	
	// Hash the key for storage
	hash := sha256.Sum256([]byte(fullKey))
	keyHash := hex.EncodeToString(hash[:])
	
	// Create prefix for display (first 16 chars + ****)
	keyPrefix := fullKey[:16] + "****"

	// Calculate expiration
	var expiresAt *time.Time
	if req.ExpiresInDays != nil && *req.ExpiresInDays > 0 {
		expiry := time.Now().AddDate(0, 0, *req.ExpiresInDays)
		expiresAt = &expiry
	}

	apiKey := &types.ApiKey{
		ID:            uuid.New().String(),
		ProjectID:     projectID,
		EnvironmentID: req.EnvironmentID,
		Name:          req.Name,
		KeyHash:       keyHash,
		KeyPrefix:     keyPrefix,
		FullKey:       fullKey, // Only included in creation response
		Permissions:   req.Permissions,
		CreatedBy:     userID,
		IsActive:      true,
		ExpiresAt:     expiresAt,
	}

	query := `
		INSERT INTO api_keys (id, project_id, environment_id, name, key_hash, key_prefix, permissions, created_by, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
	`

	_, err = r.db.ExecContext(ctx, query,
		apiKey.ID, apiKey.ProjectID, apiKey.EnvironmentID, apiKey.Name,
		apiKey.KeyHash, apiKey.KeyPrefix, pq.Array(apiKey.Permissions),
		apiKey.CreatedBy, apiKey.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	return apiKey, nil
}

// GetApiKeysByProject returns all API keys for a project
func (r *ApiKeyRepository) GetApiKeysByProject(ctx context.Context, projectID string) ([]*types.ApiKey, error) {
	query := `
		SELECT ak.id, ak.project_id, ak.environment_id, ak.name, ak.key_prefix, 
		       ak.permissions, ak.created_by, ak.created_at, ak.updated_at, 
		       ak.expires_at, ak.last_used_at, ak.is_active,
		       e.key as env_key, e.name as env_name, e.description as env_description
		FROM api_keys ak
		JOIN environments e ON ak.environment_id = e.id
		WHERE ak.project_id = $1
		ORDER BY ak.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query API keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []*types.ApiKey
	for rows.Next() {
		ak := &types.ApiKey{
			Environment: &types.Environment{},
		}

		var permissions pq.StringArray
		err := rows.Scan(
			&ak.ID, &ak.ProjectID, &ak.EnvironmentID, &ak.Name, &ak.KeyPrefix,
			&permissions, &ak.CreatedBy, &ak.CreatedAt, &ak.UpdatedAt,
			&ak.ExpiresAt, &ak.LastUsedAt, &ak.IsActive,
			&ak.Environment.Key, &ak.Environment.Name, &ak.Environment.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}

		ak.Permissions = []string(permissions)
		ak.Environment.ID = ak.EnvironmentID
		apiKeys = append(apiKeys, ak)
	}

	return apiKeys, nil
}

// AuthenticateApiKey validates an API key and returns the associated project and environment
func (r *ApiKeyRepository) AuthenticateApiKey(ctx context.Context, apiKey string) (*types.ApiKey, error) {
	// Hash the provided key
	hash := sha256.Sum256([]byte(apiKey))
	keyHash := hex.EncodeToString(hash[:])

	query := `
		SELECT ak.id, ak.project_id, ak.environment_id, ak.name, ak.key_prefix,
		       ak.permissions, ak.created_by, ak.created_at, ak.updated_at,
		       ak.expires_at, ak.last_used_at, ak.is_active,
		       e.key as env_key, e.name as env_name, e.description as env_description,
		       p.name as project_name, p.name as project_slug
		FROM api_keys ak
		JOIN environments e ON ak.environment_id = e.id
		JOIN projects p ON ak.project_id = p.id
		WHERE ak.key_hash = $1 AND ak.is_active = true
	`

	row := r.db.QueryRowContext(ctx, query, keyHash)

	ak := &types.ApiKey{
		Environment: &types.Environment{},
		Project:     &types.Project{},
	}

	var permissions pq.StringArray
	err := row.Scan(
		&ak.ID, &ak.ProjectID, &ak.EnvironmentID, &ak.Name, &ak.KeyPrefix,
		&permissions, &ak.CreatedBy, &ak.CreatedAt, &ak.UpdatedAt,
		&ak.ExpiresAt, &ak.LastUsedAt, &ak.IsActive,
		&ak.Environment.Key, &ak.Environment.Name, &ak.Environment.Description,
		&ak.Project.Name, &ak.Project.Slug,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid API key")
		}
		return nil, fmt.Errorf("failed to authenticate API key: %w", err)
	}

	ak.Permissions = []string(permissions)
	ak.Environment.ID = ak.EnvironmentID
	ak.Project.ID = ak.ProjectID

	// Check if key has expired
	if ak.ExpiresAt != nil && ak.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("API key has expired")
	}

	// Update last used timestamp
	go r.updateLastUsed(context.Background(), ak.ID)

	return ak, nil
}

// DeleteApiKey removes an API key
func (r *ApiKeyRepository) DeleteApiKey(ctx context.Context, keyID, projectID string) error {
	query := `DELETE FROM api_keys WHERE id = $1 AND project_id = $2`
	result, err := r.db.ExecContext(ctx, query, keyID, projectID)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("API key not found")
	}

	return nil
}

// UpdateApiKey updates an API key
func (r *ApiKeyRepository) UpdateApiKey(ctx context.Context, keyID string, req *types.UpdateApiKeyRequest) error {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Name != "" {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, req.Name)
		argIndex++
	}

	if len(req.Permissions) > 0 {
		setParts = append(setParts, fmt.Sprintf("permissions = $%d", argIndex))
		args = append(args, pq.Array(req.Permissions))
		argIndex++
	}

	if req.IsActive != nil {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}

	if req.ExpiresAt != nil {
		setParts = append(setParts, fmt.Sprintf("expires_at = $%d", argIndex))
		args = append(args, *req.ExpiresAt)
		argIndex++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = NOW()"))

	setClause := ""
	for i, part := range setParts {
		if i > 0 {
			setClause += ", "
		}
		setClause += part
	}

	query := fmt.Sprintf(`UPDATE api_keys SET %s WHERE id = $%d`, setClause, argIndex)
	args = append(args, keyID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update API key: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("API key not found")
	}

	return nil
}

// Helper function to get environment by ID
func (r *ApiKeyRepository) getEnvironmentByID(ctx context.Context, envID string) (*types.Environment, error) {
	query := `SELECT id, key, name, description FROM environments WHERE id = $1`
	
	env := &types.Environment{}
	err := r.db.QueryRowContext(ctx, query, envID).Scan(
		&env.ID, &env.Key, &env.Name, &env.Description,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("environment not found")
		}
		return nil, err
	}
	
	return env, nil
}

// Helper function to update last used timestamp
func (r *ApiKeyRepository) updateLastUsed(ctx context.Context, keyID string) {
	query := `UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`
	r.db.ExecContext(ctx, query, keyID)
}

// GetAllApiKeys retrieves all API keys for edge server sync
func (r *ApiKeyRepository) GetAllApiKeys(ctx context.Context) ([]*types.ApiKey, error) {
	query := `
		SELECT 
			id, project_id, environment_id, name, key_hash, key_prefix,
			permissions, created_by, created_at, expires_at, is_active
		FROM api_keys
		WHERE is_active = true
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query API keys: %w", err)
	}
	defer rows.Close()
	
	var apiKeys []*types.ApiKey
	for rows.Next() {
		apiKey := &types.ApiKey{}
		var permissionsArray pq.StringArray
		
		err := rows.Scan(
			&apiKey.ID, &apiKey.ProjectID, &apiKey.EnvironmentID,
			&apiKey.Name, &apiKey.KeyHash, &apiKey.KeyPrefix,
			&permissionsArray, &apiKey.CreatedBy, &apiKey.CreatedAt,
			&apiKey.ExpiresAt, &apiKey.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}
		
		apiKey.Permissions = []string(permissionsArray)
		apiKeys = append(apiKeys, apiKey)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating API keys: %w", err)
	}
	
	return apiKeys, nil
}
