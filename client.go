package nextengage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

// DefaultBaseURL is the production SDK endpoint. Override with WithBaseURL
// (e.g. "http://localhost:3001") for local development.
const DefaultBaseURL = "https://api.nextengage.io"

// envAPIKey is the environment variable New falls back to when no key is passed.
const envAPIKey = "NEXTENGAGE_API_KEY"

// Client is a NextEngage API client scoped to a single project (the one the API
// key belongs to). It is safe for concurrent use by multiple goroutines.
//
// Create one with New and reach the resources through its service fields:
//
//	ne, _ := nextengage.New(nextengage.WithAPIKey("ne_..."))
//	ne.Contacts.Upsert(ctx, ...)
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client

	// Resource services.
	Contacts  *ContactsService
	Segments  *SegmentsService
	Templates *TemplatesService
	Campaigns *CampaignsService

	// projectID is resolved lazily from /me and cached; mu guards it.
	mu        sync.Mutex
	projectID string
}

// Option configures a Client in New.
type Option func(*Client)

// WithAPIKey sets the project API key (ne_...). When unset, New falls back to
// the NEXTENGAGE_API_KEY environment variable.
func WithAPIKey(apiKey string) Option {
	return func(c *Client) { c.apiKey = apiKey }
}

// WithBaseURL overrides the API base URL (the SDK endpoint). Defaults to
// DefaultBaseURL. Point it at http://localhost:3001 for local development.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) { c.baseURL = baseURL }
}

// WithHTTPClient supplies a custom *http.Client (timeouts, transport, proxies).
// Defaults to http.DefaultClient.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// New creates a Client. It returns an error if no API key is available from
// either WithAPIKey or the NEXTENGAGE_API_KEY environment variable.
func New(opts ...Option) (*Client, error) {
	c := &Client{baseURL: DefaultBaseURL, httpClient: http.DefaultClient}
	for _, opt := range opts {
		opt(c)
	}
	if c.apiKey == "" {
		c.apiKey = os.Getenv(envAPIKey)
	}
	if c.apiKey == "" {
		return nil, fmt.Errorf("nextengage: missing API key — pass WithAPIKey or set the %s environment variable", envAPIKey)
	}
	c.baseURL = strings.TrimRight(c.baseURL, "/")
	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}
	c.Contacts = &ContactsService{client: c}
	c.Segments = &SegmentsService{client: c}
	c.Templates = &TemplatesService{client: c}
	c.Campaigns = &CampaignsService{client: c}
	return c, nil
}

// Error is returned for any non-2xx API response. It exposes the HTTP status,
// the server's message, and the raw response body. Callers can detect it with
// errors.As:
//
//	var apiErr *nextengage.Error
//	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
//	    // handle missing resource
//	}
type Error struct {
	// StatusCode is the HTTP status code of the response.
	StatusCode int
	// Message is the server-provided error message (or the HTTP status text).
	Message string
	// Body is the raw response body, for inspecting extra fields.
	Body []byte
}

func (e *Error) Error() string {
	return fmt.Sprintf("nextengage: HTTP %d: %s", e.StatusCode, e.Message)
}

// ProjectID resolves and caches the project id the API key is scoped to (via
// GET /me). The first successful call caches the value for the client's
// lifetime; concurrent callers share the single lookup.
func (c *Client) ProjectID(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.projectID != "" {
		return c.projectID, nil
	}
	var resp struct {
		ProjectID string `json:"projectId"`
	}
	if err := c.request(ctx, http.MethodGet, "/me", nil, &resp); err != nil {
		return "", err
	}
	if resp.ProjectID == "" {
		return "", errors.New("nextengage: /me returned no projectId — is this key a project API key?")
	}
	c.projectID = resp.ProjectID
	return c.projectID, nil
}

// request performs an HTTP call against the API. path is relative to the API
// root and must start with "/"; the "/api" prefix is added here. A non-nil body
// is JSON-encoded; a non-nil out is JSON-decoded from a 2xx response.
func (c *Client) request(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("nextengage: encode request body: %w", err)
		}
		reader = bytes.NewReader(encoded)
	}

	// API routes live under the /api prefix (baseURL is the API origin).
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+"/api"+path, reader)
	if err != nil {
		return fmt.Errorf("nextengage: build request: %w", err)
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("nextengage: request failed: %w", err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("nextengage: read response body: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return newError(res.StatusCode, res.Status, data)
	}

	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("nextengage: decode response body: %w", err)
		}
	}
	return nil
}

// projectRequest resolves the project id and dispatches to a
// /projects/{projectId}{subpath} route.
func (c *Client) projectRequest(ctx context.Context, method, subpath string, body, out any) error {
	projectID, err := c.ProjectID(ctx)
	if err != nil {
		return err
	}
	return c.request(ctx, method, "/projects/"+projectID+subpath, body, out)
}

// newError builds an *Error, extracting the server's "message" field (a string
// or, for validation failures, an array) and falling back to the status text.
func newError(status int, statusText string, body []byte) *Error {
	message := strings.TrimSpace(strings.TrimPrefix(statusText, fmt.Sprintf("%d ", status)))
	if message == "" {
		message = "request failed"
	}
	var envelope struct {
		Message json.RawMessage `json:"message"`
	}
	if json.Unmarshal(body, &envelope) == nil && len(envelope.Message) > 0 {
		var asString string
		if json.Unmarshal(envelope.Message, &asString) == nil {
			message = asString
		} else {
			// NestJS validation errors return message as a string array.
			message = string(envelope.Message)
		}
	}
	return &Error{StatusCode: status, Message: message, Body: body}
}

// Page is one page of a paginated list response.
type Page[T any] struct {
	Data   []T `json:"data"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// OKResult is the {"ok": true} acknowledgement returned by delete-style routes.
type OKResult struct {
	OK bool `json:"ok"`
}
