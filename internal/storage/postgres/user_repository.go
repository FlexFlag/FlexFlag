package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/flexflag/flexflag/pkg/types"
	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *types.User) error {
	user.ID = uuid.New().String()
	
	query := `
		INSERT INTO users (id, email, password_hash, full_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`
	
	_, err := r.db.ExecContext(ctx, query, 
		user.ID, user.Email, user.PasswordHash, user.FullName, user.Role, user.IsActive)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	return r.GetByID(ctx, user.ID, user)
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id string, user *types.User) error {
	query := `
		SELECT id, email, password_hash, full_name, role, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FullName, 
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}
	
	return nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*types.User, error) {
	user := &types.User{}
	query := `
		SELECT id, email, password_hash, full_name, role, is_active, created_at, updated_at
		FROM users
		WHERE email = $1 AND is_active = true
	`
	
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FullName, 
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	
	return user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *types.User) error {
	query := `
		UPDATE users 
		SET full_name = $2, role = $3, is_active = $4, updated_at = NOW()
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(ctx, query, user.ID, user.FullName, user.Role, user.IsActive)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	return r.GetByID(ctx, user.ID, user)
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	query := `
		UPDATE users 
		SET password_hash = $2, updated_at = NOW()
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(ctx, query, userID, passwordHash)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	return nil
}

// Delete soft deletes a user (sets is_active to false)
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE users 
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	return nil
}

// List retrieves all active users with pagination
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]types.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, role, is_active, created_at, updated_at
		FROM users
		WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()
	
	var users []types.User
	for rows.Next() {
		var user types.User
		err := rows.Scan(
			&user.ID, &user.Email, &user.PasswordHash, &user.FullName,
			&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	
	return users, nil
}

// Count returns the total number of active users
func (r *UserRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE is_active = true`
	
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	
	return count, nil
}

// EmailExists checks if an email already exists
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE email = $1 AND is_active = true`
	
	err := r.db.QueryRowContext(ctx, query, email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	
	return count > 0, nil
}