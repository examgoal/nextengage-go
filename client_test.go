package nextengage

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// newTestClient spins up an httptest server with the given handler and returns a
// client pointed at it.
func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := New(WithAPIKey("ne_test"), WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c, srv
}

func TestNewRequiresAPIKey(t *testing.T) {
	t.Setenv(envAPIKey, "")
	if _, err := New(); err == nil {
		t.Fatal("expected error when no API key is provided")
	}
	if _, err := New(WithAPIKey("ne_x")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewReadsEnvAPIKey(t *testing.T) {
	t.Setenv(envAPIKey, "ne_from_env")
	c, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if c.apiKey != "ne_from_env" {
		t.Fatalf("apiKey = %q, want ne_from_env", c.apiKey)
	}
}

func TestRequestPlumbing(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		// /me to resolve the project id, then the contact upsert.
		switch r.URL.Path {
		case "/api/me":
			if got := r.Header.Get("x-api-key"); got != "ne_test" {
				t.Errorf("x-api-key = %q, want ne_test", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"projectId": "proj_1"})
		case "/api/projects/proj_1/contacts":
			if r.Method != http.MethodPost {
				t.Errorf("method = %s, want POST", r.Method)
			}
			if ct := r.Header.Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
			body, _ := io.ReadAll(r.Body)
			var in map[string]any
			_ = json.Unmarshal(body, &in)
			if in["id"] != "user_42" {
				t.Errorf("body id = %v, want user_42", in["id"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "user_42", "email": "ada@example.com",
				"attributes": map[string]any{}, "subscriptionStatus": "subscribed",
			})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})

	got, err := c.Contacts.Upsert(context.Background(), ContactInput{ID: "user_42", Email: "ada@example.com"})
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if got.ID != "user_42" || got.Email == nil || *got.Email != "ada@example.com" {
		t.Fatalf("got %+v", got)
	}
}

func TestProjectIDCached(t *testing.T) {
	var meCalls int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/me":
			atomic.AddInt32(&meCalls, 1)
			_ = json.NewEncoder(w).Encode(map[string]string{"projectId": "proj_1"})
		default:
			_ = json.NewEncoder(w).Encode([]any{})
		}
	})

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if _, err := c.Segments.List(ctx); err != nil {
			t.Fatalf("List: %v", err)
		}
	}
	if n := atomic.LoadInt32(&meCalls); n != 1 {
		t.Fatalf("/me called %d times, want 1 (should be cached)", n)
	}
}

func TestErrorParsing(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/me" {
			_ = json.NewEncoder(w).Encode(map[string]string{"projectId": "proj_1"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"message": "Contact not found", "statusCode": 404})
	})

	_, err := c.Contacts.Get(context.Background(), "missing")
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is not *nextengage.Error: %v", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
	if apiErr.Message != "Contact not found" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "Contact not found")
	}
}

func TestAttributeOperatorsMarshal(t *testing.T) {
	cases := []struct {
		name string
		op   any
		want string
	}{
		{"increment", Increment(2), `{"$inc":2}`},
		{"arrayUnion", ArrayUnion("a", "b"), `{"$add":["a","b"]}`},
		{"arrayRemove", ArrayRemove("beta"), `{"$remove":["beta"]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.op)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(b) != tc.want {
				t.Fatalf("marshal = %s, want %s", b, tc.want)
			}
		})
	}
}

func TestNilAttributeDeletesKey(t *testing.T) {
	// A nil attribute value must serialize to JSON null (the server's "delete").
	b, err := json.Marshal(ContactInput{ID: "x", Attributes: map[string]any{"plan": nil}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string]any
	_ = json.Unmarshal(b, &decoded)
	attrs := decoded["attributes"].(map[string]any)
	if v, ok := attrs["plan"]; !ok || v != nil {
		t.Fatalf("attributes.plan = %v (ok=%v), want explicit null", v, ok)
	}
}
