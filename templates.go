package nextengage

import (
	"context"
	"net/http"
	"time"
)

// TemplatesService accesses the message templates of the API key's project. Get
// it from Client.Templates.
type TemplatesService struct {
	client *Client
}

// Template is a stored message template.
type Template struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	// Channel is "email", "push", or "sms".
	Channel string  `json:"channel"`
	Name    string  `json:"name"`
	Subject *string `json:"subject"`
	// Body is MJML or HTML for email; plain text for sms/push.
	Body string `json:"body"`
	// Variables are the template's detected interpolation variables.
	Variables []string  `json:"variables"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TemplateInput is the payload for Create.
type TemplateInput struct {
	Name    string `json:"name"`
	Subject string `json:"subject,omitempty"`
	Body    string `json:"body"`
	// Channel is "email" (default), "push", or "sms".
	Channel string `json:"channel,omitempty"`
}

// List returns the project's templates.
func (s *TemplatesService) List(ctx context.Context) ([]Template, error) {
	var out []Template
	if err := s.client.projectRequest(ctx, http.MethodGet, "/templates", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Create creates a template and returns it.
func (s *TemplatesService) Create(ctx context.Context, input TemplateInput) (*Template, error) {
	var out Template
	if err := s.client.projectRequest(ctx, http.MethodPost, "/templates", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
