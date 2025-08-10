package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flexflag/flexflag/internal/storage"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuditService struct {
	auditRepo storage.AuditRepository
}

func NewAuditService(auditRepo storage.AuditRepository) *AuditService {
	return &AuditService{
		auditRepo: auditRepo,
	}
}

// LogFlagAction logs flag-related actions (create, update, delete, toggle)
func (s *AuditService) LogFlagAction(ctx context.Context, c *gin.Context, action string, flag *types.Flag, oldFlag *types.Flag) error {
	var oldValues, newValues json.RawMessage
	var err error

	if oldFlag != nil {
		oldValues, err = json.Marshal(oldFlag)
		if err != nil {
			return err
		}
	}

	if flag != nil {
		newValues, err = json.Marshal(flag)
		if err != nil {
			return err
		}
	}

	log := &types.AuditLog{
		ID:           uuid.New().String(),
		ProjectID:    &flag.ProjectID,
		ResourceType: "flag",
		ResourceID:   flag.ID,
		Action:       action,
		OldValues:    oldValues,
		NewValues:    newValues,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.GetHeader("User-Agent"),
		CreatedAt:    time.Now(),
	}

	// Try to get user ID from context if available
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			log.UserID = &uid
		}
	}

	// Debug logging
	fmt.Printf("Creating flag audit log: %+v\n", log)
	err = s.auditRepo.Create(ctx, log)
	if err != nil {
		fmt.Printf("Flag audit log creation failed: %v\n", err)
	} else {
		fmt.Printf("Flag audit log created successfully\n")
	}
	return err
}

// LogProjectAction logs project-related actions
func (s *AuditService) LogProjectAction(ctx context.Context, c *gin.Context, action string, project *types.Project, oldProject *types.Project) error {
	var oldValues, newValues json.RawMessage
	var err error

	if oldProject != nil {
		oldValues, err = json.Marshal(oldProject)
		if err != nil {
			return err
		}
	}

	if project != nil {
		newValues, err = json.Marshal(project)
		if err != nil {
			return err
		}
	}

	log := &types.AuditLog{
		ID:           uuid.New().String(),
		ProjectID:    &project.ID,
		ResourceType: "project",
		ResourceID:   project.ID,
		Action:       action,
		OldValues:    oldValues,
		NewValues:    newValues,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.GetHeader("User-Agent"),
		CreatedAt:    time.Now(),
	}

	// Try to get user ID from context if available
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			log.UserID = &uid
		}
	}

	return s.auditRepo.Create(ctx, log)
}

// LogSegmentAction logs segment-related actions
func (s *AuditService) LogSegmentAction(ctx context.Context, c *gin.Context, action string, segment *types.Segment, oldSegment *types.Segment) error {
	var oldValues, newValues json.RawMessage
	var err error

	if oldSegment != nil {
		oldValues, err = json.Marshal(oldSegment)
		if err != nil {
			return err
		}
	}

	if segment != nil {
		newValues, err = json.Marshal(segment)
		if err != nil {
			return err
		}
	}

	log := &types.AuditLog{
		ID:           uuid.New().String(),
		ProjectID:    &segment.ProjectID,
		ResourceType: "segment",
		ResourceID:   segment.ID,
		Action:       action,
		OldValues:    oldValues,
		NewValues:    newValues,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.GetHeader("User-Agent"),
		CreatedAt:    time.Now(),
	}

	// Try to get user ID from context if available
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			log.UserID = &uid
		}
	}

	return s.auditRepo.Create(ctx, log)
}