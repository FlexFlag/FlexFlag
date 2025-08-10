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

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// Project methods

// Create creates a new project
func (r *ProjectRepository) Create(ctx context.Context, project *types.Project) error {
	project.ID = uuid.New().String()
	
	query := `
		INSERT INTO projects (id, key, name, description, created_by, is_active, settings, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`
	
	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now
	
	settingsJSON, err := json.Marshal(project.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}
	
	_, err = r.db.ExecContext(ctx, query, 
		project.ID, project.Slug, project.Name, project.Description, project.CreatedBy, 
		project.IsActive, settingsJSON)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}
	
	return nil
}

// GetByID retrieves a project by ID
func (r *ProjectRepository) GetByID(ctx context.Context, id string) (*types.Project, error) {
	query := `
		SELECT id, key, name, description, created_by, created_at, updated_at
		FROM projects
		WHERE id = $1
	`
	
	project := &types.Project{}
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&project.ID, &project.Slug, &project.Name, &project.Description,
		&project.CreatedBy, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	
	project.IsActive = true // Default for existing projects
	project.Settings = make(map[string]interface{}) // Default empty settings
	
	return project, nil
}

// GetBySlug retrieves a project by slug
func (r *ProjectRepository) GetBySlug(ctx context.Context, slug string) (*types.Project, error) {
	query := `
		SELECT id, key, name, description, created_by, created_at, updated_at
		FROM projects
		WHERE key = $1
	`
	
	project := &types.Project{}
	
	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&project.ID, &project.Slug, &project.Name, &project.Description,
		&project.CreatedBy, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	
	project.IsActive = true // Default for existing projects
	project.Settings = make(map[string]interface{}) // Default empty settings
	
	return project, nil
}

// List retrieves all projects with pagination
func (r *ProjectRepository) List(ctx context.Context, limit, offset int) ([]*types.Project, error) {
	query := `
		SELECT id, key, name, description, created_by, is_active, settings, created_at, updated_at
		FROM projects
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()
	
	var projects []*types.Project
	for rows.Next() {
		project := &types.Project{}
		var settingsJSON []byte
		var createdBy sql.NullString
		
		err := rows.Scan(
			&project.ID, &project.Slug, &project.Name, &project.Description,
			&createdBy, &project.IsActive, &settingsJSON, &project.CreatedAt, &project.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		
		// Handle nullable created_by field
		if createdBy.Valid {
			project.CreatedBy = createdBy.String
		}
		
		// Unmarshal settings JSON
		if len(settingsJSON) > 0 {
			err := json.Unmarshal(settingsJSON, &project.Settings)
			if err != nil {
				project.Settings = make(map[string]interface{}) // Default empty on error
			}
		} else {
			project.Settings = make(map[string]interface{}) // Default empty settings
		}
		
		projects = append(projects, project)
	}
	
	return projects, nil
}

// Update updates a project
func (r *ProjectRepository) Update(ctx context.Context, project *types.Project) error {
	query := `
		UPDATE projects SET
			name = $2,
			description = $3,
			is_active = $4,
			settings = $5,
			updated_at = NOW()
		WHERE id = $1
	`
	
	var settingsJSON sql.NullString
	if project.Settings != nil && len(project.Settings) > 0 {
		settingsBytes, err := json.Marshal(project.Settings)
		if err != nil {
			return fmt.Errorf("failed to marshal settings: %w", err)
		}
		settingsJSON = sql.NullString{String: string(settingsBytes), Valid: true}
	}
	
	project.UpdatedAt = time.Now()
	
	result, err := r.db.ExecContext(ctx, query,
		project.ID, project.Name, project.Description,
		project.IsActive, settingsJSON)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("project not found")
	}
	
	return nil
}

// Delete soft deletes a project
func (r *ProjectRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE projects SET is_active = false, updated_at = NOW() WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("project not found")
	}
	
	return nil
}

// SlugExists checks if a slug already exists
func (r *ProjectRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM projects WHERE key = $1`
	
	err := r.db.QueryRowContext(ctx, query, slug).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check slug existence: %w", err)
	}
	
	return count > 0, nil
}

// Environment methods

// CreateEnvironment creates a new environment for a project
func (r *ProjectRepository) CreateEnvironment(ctx context.Context, env *types.Environment) error {
	env.ID = uuid.New().String()
	
	query := `
		INSERT INTO environments (id, project_id, name, key, description, is_active, sort_order, settings, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`
	
	var settingsJSON sql.NullString
	if env.Settings != nil && len(env.Settings) > 0 {
		settingsBytes, err := json.Marshal(env.Settings)
		if err != nil {
			return fmt.Errorf("failed to marshal settings: %w", err)
		}
		settingsJSON = sql.NullString{String: string(settingsBytes), Valid: true}
	}
	
	now := time.Now()
	env.CreatedAt = now
	env.UpdatedAt = now
	
	_, err := r.db.ExecContext(ctx, query, 
		env.ID, env.ProjectID, env.Name, env.Key, env.Description,
		env.IsActive, env.SortOrder, settingsJSON)
	if err != nil {
		return fmt.Errorf("failed to create environment: %w", err)
	}
	
	return nil
}

// GetEnvironmentsByProject retrieves all environments for a project
func (r *ProjectRepository) GetEnvironmentsByProject(ctx context.Context, projectID string) ([]*types.Environment, error) {
	query := `
		SELECT id, project_id, name, key, description, is_active, sort_order, settings, created_at, updated_at
		FROM environments
		WHERE project_id = $1 AND is_active = true
		ORDER BY sort_order, name
	`
	
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list environments: %w", err)
	}
	defer rows.Close()
	
	var environments []*types.Environment
	for rows.Next() {
		env := &types.Environment{}
		var settingsJSON sql.NullString
		
		err := rows.Scan(
			&env.ID, &env.ProjectID, &env.Name, &env.Key, &env.Description,
			&env.IsActive, &env.SortOrder, &settingsJSON,
			&env.CreatedAt, &env.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan environment: %w", err)
		}
		
		if settingsJSON.Valid {
			if err := json.Unmarshal([]byte(settingsJSON.String), &env.Settings); err != nil {
				return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
			}
		}
		
		environments = append(environments, env)
	}
	
	return environments, nil
}