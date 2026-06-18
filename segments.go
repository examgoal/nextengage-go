package nextengage

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

// SegmentsService accesses the segments of the API key's project. Get it from
// Client.Segments.
type SegmentsService struct {
	client *Client
}

// Segment is a stored audience segment.
type Segment struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Name      string `json:"name"`
	// Type is "static" or "dynamic".
	Type string `json:"type"`
	// Filter is the dynamic segment's rule tree (null for static segments). It is
	// left raw so callers can decode it into their own shape if needed.
	Filter            json.RawMessage `json:"filter,omitempty"`
	MaterializedCount int             `json:"materializedCount"`
	LastRefreshedAt   *time.Time      `json:"lastRefreshedAt"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
}

// RefreshResult is returned by Refresh: the number of contacts now in the
// segment.
type RefreshResult struct {
	Count int `json:"count"`
}

// List returns the project's segments.
func (s *SegmentsService) List(ctx context.Context) ([]Segment, error) {
	var out []Segment
	if err := s.client.projectRequest(ctx, http.MethodGet, "/segments", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Get fetches a single segment by id.
func (s *SegmentsService) Get(ctx context.Context, id string) (*Segment, error) {
	var out Segment
	if err := s.client.projectRequest(ctx, http.MethodGet, "/segments/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Refresh recomputes a dynamic segment's membership and returns the new count.
func (s *SegmentsService) Refresh(ctx context.Context, id string) (*RefreshResult, error) {
	var out RefreshResult
	if err := s.client.projectRequest(ctx, http.MethodPost, "/segments/"+url.PathEscape(id)+"/refresh", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
