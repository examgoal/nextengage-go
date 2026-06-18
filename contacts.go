package nextengage

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// ContactsService accesses the contacts of the API key's project. Get it from
// Client.Contacts.
type ContactsService struct {
	client *Client
}

// Contact is a stored contact.
type Contact struct {
	// ID is the contact's id — your own user id (it IS the primary key).
	ID                 string         `json:"id"`
	Name               *string        `json:"name"`
	Email              *string        `json:"email"`
	Phone              *string        `json:"phone"`
	Attributes         map[string]any `json:"attributes"`
	SubscriptionStatus string         `json:"subscriptionStatus"`
}

// PushToken is a device push token tagged with the Firebase project it was
// registered with, so push campaigns route it to the right credential.
type PushToken struct {
	Token string `json:"token"`
	// Project is the Firebase project_id the token was registered with.
	Project string `json:"project,omitempty"`
}

// ContactInput is the payload for Upsert and BulkImport.
type ContactInput struct {
	// ID is YOUR own unique id for the user. It IS the contact's key: an upsert
	// with the same id updates the same contact, so always pass a stable id you
	// control. Required.
	ID    string `json:"id"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
	// Attributes merge into the contact's existing attributes (top-level): keys
	// you send are set, keys you omit are kept, and a key set to nil is deleted.
	// A value may also be an operator — Increment, ArrayUnion, ArrayRemove — that
	// mutates the existing value atomically.
	Attributes map[string]any `json:"attributes,omitempty"`
	// PushTokens merge into the contact's existing set (deduped by token), so each
	// device can register on its own without wiping the others. Send an empty
	// non-nil slice to clear all.
	PushTokens []PushToken `json:"pushTokens,omitempty"`
	// SubscriptionStatus is one of "subscribed", "unsubscribed", "bounced",
	// "complained".
	SubscriptionStatus string `json:"subscriptionStatus,omitempty"`
}

// ContactUpdate is the partial payload for Update. Omitted fields are left
// unchanged.
type ContactUpdate struct {
	Name               string         `json:"name,omitempty"`
	Email              string         `json:"email,omitempty"`
	Phone              string         `json:"phone,omitempty"`
	Attributes         map[string]any `json:"attributes,omitempty"`
	SubscriptionStatus string         `json:"subscriptionStatus,omitempty"`
}

// ListContactsParams are the optional filters for List.
type ListContactsParams struct {
	Limit  int
	Offset int
	Search string
}

// Event is a behavioral event for Track.
type Event struct {
	// ID is the contact id (= your own user id) the event belongs to. Required.
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Properties map[string]any `json:"properties,omitempty"`
}

// TrackResult is returned by Track.
type TrackResult struct {
	OK        bool   `json:"ok"`
	ContactID string `json:"contactId"`
}

// BulkImportResult is returned by BulkImport.
type BulkImportResult struct {
	Imported int `json:"imported"`
	NewCount int `json:"newCount"`
}

// List returns a page of contacts. Pass nil params for defaults.
func (s *ContactsService) List(ctx context.Context, params *ListContactsParams) (*Page[Contact], error) {
	q := url.Values{}
	if params != nil {
		if params.Limit > 0 {
			q.Set("limit", strconv.Itoa(params.Limit))
		}
		if params.Offset > 0 {
			q.Set("offset", strconv.Itoa(params.Offset))
		}
		if params.Search != "" {
			q.Set("search", params.Search)
		}
	}
	path := "/contacts"
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var out Page[Contact]
	if err := s.client.projectRequest(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a single contact by its id (your own user id).
func (s *ContactsService) Get(ctx context.Context, id string) (*Contact, error) {
	var out Contact
	if err := s.client.projectRequest(ctx, http.MethodGet, "/contacts/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Upsert creates or updates a contact (keyed by its ID). It is persisted
// synchronously and returns the saved Contact.
func (s *ContactsService) Upsert(ctx context.Context, input ContactInput) (*Contact, error) {
	var out Contact
	if err := s.client.projectRequest(ctx, http.MethodPost, "/contacts", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update patches an existing contact by id. Omitted fields are left unchanged.
func (s *ContactsService) Update(ctx context.Context, id string, input ContactUpdate) (*Contact, error) {
	var out Contact
	if err := s.client.projectRequest(ctx, http.MethodPatch, "/contacts/"+url.PathEscape(id), input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// BulkImport imports many contacts at once (each keyed by its ID). It is
// persisted synchronously and returns counts of imported and newly-created
// contacts.
func (s *ContactsService) BulkImport(ctx context.Context, contacts []ContactInput) (*BulkImportResult, error) {
	var out BulkImportResult
	body := map[string]any{"contacts": contacts}
	if err := s.client.projectRequest(ctx, http.MethodPost, "/contacts/bulk", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Track records a behavioral event (feeds dynamic segments). It is persisted
// synchronously.
func (s *ContactsService) Track(ctx context.Context, event Event) (*TrackResult, error) {
	var out TrackResult
	if err := s.client.projectRequest(ctx, http.MethodPost, "/contacts/events", event, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a contact by its id (your own user id).
func (s *ContactsService) Delete(ctx context.Context, id string) (*OKResult, error) {
	var out OKResult
	if err := s.client.projectRequest(ctx, http.MethodDelete, "/contacts/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
