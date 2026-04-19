package handlers

import (
	"encoding/json"
	"net/http"

	approvalPkg "github.com/flexflag/flexflag/internal/approval"
	"github.com/flexflag/flexflag/internal/storage"
	"github.com/flexflag/flexflag/internal/storage/postgres"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
)

// ApprovalHandler manages flag-change approval requests.
type ApprovalHandler struct {
	approvalRepo *postgres.ApprovalRepository
	flagRepo     storage.FlagRepository
	projectRepo  *postgres.ProjectRepository
	uiBaseURL    string // e.g. "http://localhost:3000"
}

func NewApprovalHandler(
	approvalRepo *postgres.ApprovalRepository,
	flagRepo storage.FlagRepository,
	projectRepo *postgres.ProjectRepository,
	uiBaseURL string,
) *ApprovalHandler {
	return &ApprovalHandler{
		approvalRepo: approvalRepo,
		flagRepo:     flagRepo,
		projectRepo:  projectRepo,
		uiBaseURL:    uiBaseURL,
	}
}

// ListApprovals returns approval requests for a project.
// GET /approvals?project_id=xxx&status=pending
func (h *ApprovalHandler) ListApprovals(c *gin.Context) {
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}
	status := types.ApprovalStatus(c.Query("status"))

	reqs, err := h.approvalRepo.ListByProject(c.Request.Context(), projectID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if reqs == nil {
		reqs = []*types.ApprovalRequest{}
	}
	c.JSON(http.StatusOK, gin.H{"approvals": reqs})
}

// GetPendingCount returns the count of pending approvals for a project.
// GET /approvals/count?project_id=xxx
func (h *ApprovalHandler) GetPendingCount(c *gin.Context) {
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}
	count, err := h.approvalRepo.PendingCount(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

// ApproveRequest approves a pending request and applies the flag change.
// POST /approvals/:id/approve
func (h *ApprovalHandler) ApproveRequest(c *gin.Context) {
	id := c.Param("id")

	var body types.ReviewRequest
	_ = c.ShouldBindJSON(&body)

	req, err := h.approvalRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "approval request not found"})
		return
	}
	if req.Status != types.ApprovalStatusPending {
		c.JSON(http.StatusConflict, gin.H{"error": "request is no longer pending"})
		return
	}

	reviewerID := ""
	if userID, exists := c.Get("userID"); exists {
		reviewerID, _ = userID.(string)
	}

	// Apply the change
	if err := h.applyChange(c, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to apply change: " + err.Error()})
		return
	}

	if err := h.approvalRepo.UpdateStatus(c.Request.Context(), id, types.ApprovalStatusApproved, reviewerID, body.Note); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "approved and applied"})
}

// RejectRequest rejects a pending request without applying any change.
// POST /approvals/:id/reject
func (h *ApprovalHandler) RejectRequest(c *gin.Context) {
	id := c.Param("id")

	var body types.ReviewRequest
	_ = c.ShouldBindJSON(&body)

	req, err := h.approvalRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "approval request not found"})
		return
	}
	if req.Status != types.ApprovalStatusPending {
		c.JSON(http.StatusConflict, gin.H{"error": "request is no longer pending"})
		return
	}

	reviewerID := ""
	if userID, exists := c.Get("userID"); exists {
		reviewerID, _ = userID.(string)
	}

	if err := h.approvalRepo.UpdateStatus(c.Request.Context(), id, types.ApprovalStatusRejected, reviewerID, body.Note); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "rejected"})
}

// applyChange executes the pending flag change stored in the approval payload.
func (h *ApprovalHandler) applyChange(c *gin.Context, req *types.ApprovalRequest) error {
	flag, err := h.flagRepo.GetByProjectKey(c.Request.Context(), req.ProjectID, req.FlagKey, req.Environment)
	if err != nil {
		return err
	}

	switch req.ChangeType {
	case types.ChangeTypeToggle:
		if val, ok := req.Payload["new_enabled"].(bool); ok {
			flag.Enabled = val
		} else {
			flag.Enabled = !flag.Enabled
		}
		return h.flagRepo.Update(c.Request.Context(), flag)

	case types.ChangeTypeUpdate:
		// Payload contains the full updated flag fields as JSON
		if raw, ok := req.Payload["flag_update"]; ok {
			b, _ := json.Marshal(raw)
			if err := json.Unmarshal(b, flag); err == nil {
				return h.flagRepo.Update(c.Request.Context(), flag)
			}
		}
		return nil

	case types.ChangeTypeDelete:
		return h.flagRepo.Delete(c.Request.Context(), req.FlagKey, req.Environment)
	}
	return nil
}

// UpdateProjectApprovalSettings saves approval workflow settings for a project.
// PUT /approvals/settings?project_id=xxx
func (h *ApprovalHandler) UpdateProjectApprovalSettings(c *gin.Context) {
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	var body struct {
		RequireApproval bool   `json:"require_approval"`
		SlackWebhookURL string `json:"slack_webhook_url"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectRepo.GetByID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	if project.Settings == nil {
		project.Settings = map[string]interface{}{}
	}
	project.Settings["require_approval"] = body.RequireApproval
	project.Settings["slack_webhook_url"] = body.SlackWebhookURL

	if err := h.projectRepo.Update(c.Request.Context(), project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"require_approval":  body.RequireApproval,
		"slack_webhook_url": body.SlackWebhookURL,
	})
}

// GetProjectApprovalSettings returns the approval settings for a project.
// GET /approvals/settings?project_id=xxx
func (h *ApprovalHandler) GetProjectApprovalSettings(c *gin.Context) {
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	project, err := h.projectRepo.GetByID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	requireApproval, _ := project.Settings["require_approval"].(bool)
	slackWebhookURL, _ := project.Settings["slack_webhook_url"].(string)

	c.JSON(http.StatusOK, gin.H{
		"require_approval":  requireApproval,
		"slack_webhook_url": slackWebhookURL,
	})
}

// ── helper used by FlagHandler ────────────────────────────────────────────────

// CreateApprovalRequest creates an approval record and fires the Slack notification.
// Called by FlagHandler when a project has require_approval enabled.
func CreateApprovalRequest(
	c *gin.Context,
	approvalRepo *postgres.ApprovalRepository,
	projectRepo *postgres.ProjectRepository,
	projectID, flagKey, environment string,
	changeType types.ChangeType,
	payload map[string]interface{},
	uiBaseURL string,
) (*types.ApprovalRequest, error) {
	// Get requester identity from JWT claims
	requesterID := ""
	requesterName := ""
	if uid, ok := c.Get("userID"); ok {
		requesterID, _ = uid.(string)
	}
	if uname, ok := c.Get("userName"); ok {
		requesterName, _ = uname.(string)
	}
	if uname, ok := c.Get("userEmail"); ok && requesterName == "" {
		requesterName, _ = uname.(string)
	}

	req := &types.ApprovalRequest{
		ProjectID:       projectID,
		FlagKey:         flagKey,
		Environment:     environment,
		ChangeType:      changeType,
		Payload:         payload,
		RequestedBy:     requesterID,
		RequestedByName: requesterName,
		Status:          types.ApprovalStatusPending,
	}

	if err := approvalRepo.Create(c.Request.Context(), req); err != nil {
		return nil, err
	}

	// Fire Slack notification in background — non-fatal
	go func() {
		project, err := projectRepo.GetByID(c.Request.Context(), projectID)
		if err != nil {
			return
		}
		webhookURL, _ := project.Settings["slack_webhook_url"].(string)
		_ = approvalPkg.SendSlackNotification(webhookURL, uiBaseURL, req)
	}()

	return req, nil
}
