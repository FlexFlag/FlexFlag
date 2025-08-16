package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return db, mock
}

func TestNewFlagRepository(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()

	repo := NewFlagRepository(db)
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestFlagRepository_Create_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewFlagRepository(db)
	ctx := context.Background()

	flag := &types.Flag{
		ID:          uuid.New().String(),
		ProjectID:   "proj_123",
		Key:         "test-flag",
		Name:        "Test Flag",
		Description: "A test flag",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "production",
		Tags:        []string{"test", "feature"},
		Metadata:    map[string]interface{}{"team": "backend"},
	}

	mock.ExpectExec("INSERT INTO flags").
		WithArgs(
			flag.ID,
			flag.ProjectID,
			flag.Key,
			flag.Name,
			flag.Description,
			flag.Type,
			flag.Enabled,
			flag.Default,
			sqlmock.AnyArg(), // variations
			sqlmock.AnyArg(), // targeting
			flag.Environment,
			pq.Array(flag.Tags),
			sqlmock.AnyArg(), // metadata
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(ctx, flag)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFlagRepository_GetByKey_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewFlagRepository(db)
	ctx := context.Background()

	expectedFlag := &types.Flag{
		ID:          uuid.New().String(),
		ProjectID:   "proj_123",
		Key:         "test-flag",
		Name:        "Test Flag",
		Description: "A test flag",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "production",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	rows := sqlmock.NewRows([]string{
		"id", "project_id", "key", "name", "description", "type", "enabled",
		"default_value", "variations", "targeting", "environment",
		"tags", "metadata", "created_at", "updated_at",
	}).AddRow(
		expectedFlag.ID,
		expectedFlag.ProjectID,
		expectedFlag.Key,
		expectedFlag.Name,
		expectedFlag.Description,
		expectedFlag.Type,
		expectedFlag.Enabled,
		expectedFlag.Default,
		[]byte("[]"), // variations (empty JSON array)
		nil, // targeting
		expectedFlag.Environment,
		pq.Array([]string{}), // tags
		nil, // metadata
		expectedFlag.CreatedAt,
		expectedFlag.UpdatedAt,
	)

	mock.ExpectQuery("SELECT (.+) FROM flags WHERE key = \\$1 AND environment = \\$2").
		WithArgs("test-flag", "production").
		WillReturnRows(rows)

	flag, err := repo.GetByKey(ctx, "test-flag", "production")
	
	assert.NoError(t, err)
	assert.NotNil(t, flag)
	assert.Equal(t, expectedFlag.Key, flag.Key)
	assert.Equal(t, expectedFlag.Type, flag.Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFlagRepository_GetByKey_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewFlagRepository(db)
	ctx := context.Background()

	mock.ExpectQuery("SELECT (.+) FROM flags WHERE key = \\$1 AND environment = \\$2").
		WithArgs("non-existent", "production").
		WillReturnError(sql.ErrNoRows)

	flag, err := repo.GetByKey(ctx, "non-existent", "production")
	
	assert.Error(t, err)
	assert.Nil(t, flag)
	assert.Contains(t, err.Error(), "flag not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFlagRepository_Update_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewFlagRepository(db)
	ctx := context.Background()

	flag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "test-flag",
		Name:        "Updated Flag",
		Description: "Updated description",
		Type:        types.FlagTypeString,
		Enabled:     false,
		Default:     json.RawMessage(`"updated"`),
		Environment: "production",
	}

	mock.ExpectExec("UPDATE flags SET").
		WithArgs(
			flag.Key,
			flag.Environment,
			flag.Name,
			flag.Description,
			flag.Type,
			flag.Enabled,
			flag.Default,
			sqlmock.AnyArg(), // variations
			sqlmock.AnyArg(), // targeting
			sqlmock.AnyArg(), // tags
			sqlmock.AnyArg(), // metadata
			sqlmock.AnyArg(), // updated_at
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(ctx, flag)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFlagRepository_Delete_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewFlagRepository(db)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM flags WHERE key = \\$1 AND environment = \\$2").
		WithArgs("test-flag", "production").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(ctx, "test-flag", "production")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFlagRepository_List_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewFlagRepository(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"id", "project_id", "key", "name", "description", "type", "enabled",
		"default_value", "variations", "targeting", "environment",
		"tags", "metadata", "created_at", "updated_at",
	}).
		AddRow(
			"flag_1", "proj_123", "flag-1", "Flag 1", "First flag",
			types.FlagTypeBoolean, true, json.RawMessage(`true`),
			[]byte("[]"), nil, "production",
			pq.Array([]string{"test"}), nil, time.Now(), time.Now(),
		).
		AddRow(
			"flag_2", "proj_123", "flag-2", "Flag 2", "Second flag",
			types.FlagTypeString, false, json.RawMessage(`"default"`),
			[]byte("[]"), nil, "production",
			pq.Array([]string{"feature"}), nil, time.Now(), time.Now(),
		)

	mock.ExpectQuery("SELECT (.+) FROM flags WHERE environment = \\$1").
		WithArgs("production").
		WillReturnRows(rows)

	flags, err := repo.List(ctx, "production")
	
	assert.NoError(t, err)
	assert.Len(t, flags, 2)
	assert.Equal(t, "flag-1", flags[0].Key)
	assert.Equal(t, "flag-2", flags[1].Key)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFlagRepository_ListByProject_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewFlagRepository(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"id", "project_id", "key", "name", "description", "type", "enabled",
		"default_value", "variations", "targeting", "environment",
		"tags", "metadata", "created_at", "updated_at",
	}).
		AddRow(
			"flag_1", "proj_123", "project-flag", "Project Flag", "A project flag",
			types.FlagTypeNumber, true, json.RawMessage(`42`),
			[]byte("[]"), nil, "staging",
			pq.Array([]string{}), nil, time.Now(), time.Now(),
		)

	mock.ExpectQuery("SELECT (.+) FROM flags WHERE project_id = \\$1 AND environment = \\$2").
		WithArgs("proj_123", "staging").
		WillReturnRows(rows)

	flags, err := repo.ListByProject(ctx, "proj_123", "staging")
	
	assert.NoError(t, err)
	assert.Len(t, flags, 1)
	assert.Equal(t, "project-flag", flags[0].Key)
	assert.Equal(t, "proj_123", flags[0].ProjectID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFlagRepository_ListByTags_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewFlagRepository(db)
	ctx := context.Background()

	tags := []string{"feature", "beta"}

	rows := sqlmock.NewRows([]string{
		"id", "project_id", "key", "name", "description", "type", "enabled",
		"default_value", "variations", "targeting", "environment",
		"tags", "metadata", "created_at", "updated_at",
	}).
		AddRow(
			"flag_1", "proj_123", "tagged-flag", "Tagged Flag", "A tagged flag",
			types.FlagTypeJSON, true, json.RawMessage(`{}`),
			[]byte("[]"), nil, "production",
			pq.Array([]string{"feature", "beta"}), nil, time.Now(), time.Now(),
		)

	mock.ExpectQuery("SELECT (.+) FROM flags WHERE environment = \\$1 AND tags && \\$2").
		WithArgs("production", pq.Array(tags)).
		WillReturnRows(rows)

	flags, err := repo.ListByTags(ctx, "production", tags)
	
	assert.NoError(t, err)
	assert.Len(t, flags, 1)
	assert.Equal(t, "tagged-flag", flags[0].Key)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNewProjectRepository(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()

	repo := NewProjectRepository(db)
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestNewApiKeyRepository(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()

	repo := NewApiKeyRepository(db)
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestNewAuditRepository(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()

	repo := NewAuditRepository(db)
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestAuditRepository_Create_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewAuditRepository(db)
	ctx := context.Background()

	projectID := "proj_123"
	userID := "user_456"
	
	log := &types.AuditLog{
		ID:           uuid.New().String(),
		ProjectID:    &projectID,
		UserID:       &userID,
		ResourceType: "flag",
		ResourceID:   "flag_789",
		Action:       "update",
		OldValues:    json.RawMessage(`{"enabled":false}`),
		NewValues:    json.RawMessage(`{"enabled":true}`),
		Metadata:     map[string]interface{}{"source": "api"},
		IPAddress:    "192.168.1.1",
		UserAgent:    "Mozilla/5.0",
	}

	mock.ExpectExec("INSERT INTO audit_logs").
		WithArgs(
			log.ProjectID,
			log.UserID,
			log.ResourceType,
			log.ResourceID,
			log.Action,
			sqlmock.AnyArg(), // old_values 
			sqlmock.AnyArg(), // new_values
			sqlmock.AnyArg(), // metadata
			log.IPAddress,
			log.UserAgent,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(ctx, log)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNewRolloutRepository(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()

	repo := NewRolloutRepository(db)
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestNewSegmentRepository(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()

	repo := NewSegmentRepository(db)
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestSegmentRepository_Create_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewSegmentRepository(db)
	ctx := context.Background()

	segment := &types.Segment{
		ID:          uuid.New().String(),
		ProjectID:   "proj_123",
		Key:         "premium-users",
		Name:        "Premium Users",
		Description: "Users with premium subscription",
		Rules: []types.TargetingRule{
			{
				ID:        "rule_1",
				Attribute: "tier",
				Operator:  "eq",
				Values:    []string{"premium"},
			},
		},
		Environment: "production",
	}

	mock.ExpectExec("INSERT INTO segments").
		WithArgs(
			sqlmock.AnyArg(), // id (UUID)
			segment.ProjectID,
			segment.Key,
			segment.Name,
			segment.Description,
			sqlmock.AnyArg(), // rules
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(ctx, segment)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSegmentRepository_GetByKey_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewSegmentRepository(db)
	ctx := context.Background()

	expectedSegment := &types.Segment{
		ID:          uuid.New().String(),
		ProjectID:   "proj_123",
		Key:         "premium-users",
		Name:        "Premium Users",
		Description: "Premium segment",
		Environment: "production",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	rows := sqlmock.NewRows([]string{
		"id", "project_id", "key", "name", "description", "rules",
		"created_at", "updated_at",
	}).AddRow(
		expectedSegment.ID,
		expectedSegment.ProjectID,
		expectedSegment.Key,
		expectedSegment.Name,
		expectedSegment.Description,
		[]byte("[]"), // rules
		expectedSegment.CreatedAt,
		expectedSegment.UpdatedAt,
	)

	mock.ExpectQuery("SELECT (.+) FROM segments WHERE project_id = \\$1 AND key = \\$2").
		WithArgs("proj_123", "premium-users").
		WillReturnRows(rows)

	segment, err := repo.GetByKey(ctx, "proj_123", "premium-users")
	
	assert.NoError(t, err)
	assert.NotNil(t, segment)
	assert.Equal(t, expectedSegment.Key, segment.Key)
	assert.NoError(t, mock.ExpectationsWereMet())
}