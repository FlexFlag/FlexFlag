package storage

import (
	"context"

	"github.com/flexflag/flexflag/pkg/types"
)

type FlagRepository interface {
	Create(ctx context.Context, flag *types.Flag) error
	GetByKey(ctx context.Context, key, environment string) (*types.Flag, error)
	GetByProjectKey(ctx context.Context, projectID, key, environment string) (*types.Flag, error)
	Update(ctx context.Context, flag *types.Flag) error
	Delete(ctx context.Context, key, environment string) error
	List(ctx context.Context, environment string) ([]*types.Flag, error)
	ListByProject(ctx context.Context, projectID, environment string) ([]*types.Flag, error)
	ListByTags(ctx context.Context, environment string, tags []string) ([]*types.Flag, error)
}

type SegmentRepository interface {
	Create(ctx context.Context, segment *types.Segment) error
	GetByKey(ctx context.Context, key, environment string) (*types.Segment, error)
	Update(ctx context.Context, segment *types.Segment) error
	Delete(ctx context.Context, key, environment string) error
	List(ctx context.Context, environment string) ([]*types.Segment, error)
}

type EvaluationRepository interface {
	RecordEvaluation(ctx context.Context, req *types.EvaluationRequest, resp *types.EvaluationResponse) error
	GetEvaluationStats(ctx context.Context, flagKey, environment string) (*EvaluationStats, error)
}

type EvaluationStats struct {
	FlagKey       string            `json:"flag_key"`
	Environment   string            `json:"environment"`
	TotalCount    int64             `json:"total_count"`
	VariationCounts map[string]int64 `json:"variation_counts"`
	LastEvaluated *types.EvaluationResponse `json:"last_evaluated"`
}

type RolloutRepository interface {
	Create(ctx context.Context, rollout *types.Rollout) error
	GetByID(ctx context.Context, id string) (*types.Rollout, error)
	GetByFlag(ctx context.Context, flagID, environment string) ([]*types.Rollout, error)
	GetByProject(ctx context.Context, projectID, environment string) ([]*types.Rollout, error)
	GetActiveRollouts(ctx context.Context, flagID, environment string) ([]*types.Rollout, error)
	Update(ctx context.Context, rollout *types.Rollout) error
	Delete(ctx context.Context, id string) error
	CreateStickyAssignment(ctx context.Context, assignment *types.StickyAssignment) error
	GetStickyAssignment(ctx context.Context, flagID, environment, userKey string) (*types.StickyAssignment, error)
	DeleteStickyAssignment(ctx context.Context, flagID, environment, userKey string) error
	CleanupExpiredAssignments(ctx context.Context) error
}

type AuditRepository interface {
	Create(ctx context.Context, log *types.AuditLog) error
	List(ctx context.Context, projectID string, limit, offset int) ([]*types.AuditLog, error)
	ListByResource(ctx context.Context, resourceType, resourceID string, limit, offset int) ([]*types.AuditLog, error)
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*types.AuditLog, error)
}

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl int) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}