// Package approval provides the Slack notification sender for approval requests.
package approval

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/flexflag/flexflag/pkg/types"
)

// SendSlackNotification posts an approval request card to the configured webhook.
// baseURL is the FlexFlag UI origin (e.g. "http://localhost:3000") used to build the review link.
func SendSlackNotification(webhookURL, baseURL string, req *types.ApprovalRequest) error {
	if webhookURL == "" {
		return nil
	}

	changeDesc := describeChange(req)
	reviewURL := fmt.Sprintf("%s/projects/%s/approvals", strings.TrimRight(baseURL, "/"), req.ProjectID)

	payload := map[string]interface{}{
		"blocks": []map[string]interface{}{
			{
				"type": "header",
				"text": map[string]string{
					"type": "plain_text",
					"text": "Flag Change Needs Approval",
				},
			},
			{
				"type": "section",
				"fields": []map[string]string{
					{"type": "mrkdwn", "text": fmt.Sprintf("*Flag:*\n`%s`", req.FlagKey)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Environment:*\n%s", req.Environment)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Change:*\n%s", changeDesc)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Requested by:*\n%s", requestedByName(req))},
				},
			},
			{
				"type": "context",
				"elements": []map[string]string{
					{"type": "mrkdwn", "text": fmt.Sprintf("Requested at %s", req.CreatedAt.UTC().Format(time.RFC822))},
				},
			},
			{
				"type": "actions",
				"elements": []map[string]interface{}{
					{
						"type": "button",
						"style": "primary",
						"text":  map[string]string{"type": "plain_text", "text": "Review in FlexFlag"},
						"url":   reviewURL,
					},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal slack payload: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(body)) //nolint:noctx
	if err != nil {
		return fmt.Errorf("slack webhook post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("slack webhook returned %d", resp.StatusCode)
	}
	return nil
}

func describeChange(req *types.ApprovalRequest) string {
	switch req.ChangeType {
	case types.ChangeTypeToggle:
		// payload has "current_enabled" bool
		if cur, ok := req.Payload["current_enabled"].(bool); ok {
			if cur {
				return "Disable flag (currently enabled)"
			}
			return "Enable flag (currently disabled)"
		}
		return "Toggle flag"
	case types.ChangeTypeUpdate:
		return "Update flag configuration"
	case types.ChangeTypeDelete:
		return "Delete flag"
	default:
		return string(req.ChangeType)
	}
}

func requestedByName(req *types.ApprovalRequest) string {
	if req.RequestedByName != "" {
		return req.RequestedByName
	}
	if req.RequestedBy != "" {
		return req.RequestedBy
	}
	return "Unknown"
}
