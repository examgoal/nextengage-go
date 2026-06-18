package nextengage

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// CampaignsService accesses the campaigns of the API key's project. Get it from
// Client.Campaigns.
type CampaignsService struct {
	client *Client
}

// Campaign is a stored campaign.
type Campaign struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Name      string `json:"name"`
	// Channel is "email", "push", or "sms".
	Channel string `json:"channel"`
	// Status is the campaign lifecycle state, e.g. "draft", "scheduled",
	// "sending", "sent", "cancelled".
	Status             string         `json:"status"`
	SegmentID          *string        `json:"segmentId"`
	TemplateID         *string        `json:"templateId"`
	CredentialID       *string        `json:"credentialId"`
	FromName           *string        `json:"fromName"`
	FromEmail          *string        `json:"fromEmail"`
	ScheduledAt        *time.Time     `json:"scheduledAt"`
	Recurrence         *string        `json:"recurrence"`
	MaxSendsPerContact *int           `json:"maxSendsPerContact"`
	DailySendLimit     *int           `json:"dailySendLimit"`
	PushTopic          *string        `json:"pushTopic"`
	SentAt             *time.Time     `json:"sentAt"`
	TotalSent          int            `json:"totalSent"`
	Stats              map[string]any `json:"stats"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
}

// CampaignInput is the payload for Create.
type CampaignInput struct {
	Name string `json:"name"`
	// Channel is "email" (default), "push", or "sms".
	Channel    string `json:"channel,omitempty"`
	SegmentID  string `json:"segmentId,omitempty"`
	TemplateID string `json:"templateId,omitempty"`
	FromName   string `json:"fromName,omitempty"`
	FromEmail  string `json:"fromEmail,omitempty"`
	// ScheduledAt is an RFC 3339 timestamp; when set, the campaign is scheduled
	// rather than sent immediately.
	ScheduledAt string `json:"scheduledAt,omitempty"`
	// PushTopic broadcasts a push campaign to an FCM topic; when set the campaign
	// needs no segment.
	PushTopic string `json:"pushTopic,omitempty"`
}

// SendResult is returned by Send.
type SendResult struct {
	OK     bool   `json:"ok"`
	Status string `json:"status"`
}

// List returns the project's campaigns.
func (s *CampaignsService) List(ctx context.Context) ([]Campaign, error) {
	var out []Campaign
	if err := s.client.projectRequest(ctx, http.MethodGet, "/campaigns", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Create creates a campaign and returns it.
func (s *CampaignsService) Create(ctx context.Context, input CampaignInput) (*Campaign, error) {
	var out Campaign
	if err := s.client.projectRequest(ctx, http.MethodPost, "/campaigns", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Send dispatches a campaign and returns its new status.
func (s *CampaignsService) Send(ctx context.Context, id string) (*SendResult, error) {
	var out SendResult
	if err := s.client.projectRequest(ctx, http.MethodPost, "/campaigns/"+url.PathEscape(id)+"/send", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
