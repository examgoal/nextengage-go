# NextEngage Go SDK

Official Go SDK for the NextEngage marketing platform — programmatic access to a
single project (contacts, events, segments, templates, campaigns) using a
project API key. The Go counterpart of [`@nextengage/sdk`](https://www.npmjs.com/package/@nextengage/sdk).

[![Go Reference](https://pkg.go.dev/badge/github.com/examgoal/nextengage-go.svg)](https://pkg.go.dev/github.com/examgoal/nextengage-go)

## Install

```bash
go get github.com/examgoal/nextengage-go
```

Requires Go 1.22+.

## Usage

```go
package main

import (
	"context"
	"log"

	nextengage "github.com/examgoal/nextengage-go"
)

func main() {
	// apiKey falls back to the NEXTENGAGE_API_KEY env var, so WithAPIKey is optional:
	ne, err := nextengage.New()
	if err != nil {
		log.Fatal(err)
	}

	// …or pass it explicitly (and optionally override the endpoint for local dev):
	// ne, err := nextengage.New(
	// 	nextengage.WithAPIKey("ne_..."),
	// 	nextengage.WithBaseURL("http://localhost:3001"),
	// )

	ctx := context.Background()

	// Upsert a contact, keyed by YOUR own user id — persisted synchronously,
	// returns the saved Contact. Attributes MERGE: this sets country/plan without
	// touching other attributes; set a key to nil to delete it.
	_, err = ne.Contacts.Upsert(ctx, nextengage.ContactInput{
		ID:         "user_42",
		Email:      "ada@example.com",
		Attributes: map[string]any{"country": "IN", "plan": "pro"},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Track a behavioral event (feeds dynamic segments).
	_, _ = ne.Contacts.Track(ctx, nextengage.Event{
		ID: "user_42", Name: "purchased", Properties: map[string]any{"amount": 49},
	})

	// Bulk import (each row keyed by its id).
	_, _ = ne.Contacts.BulkImport(ctx, []nextengage.ContactInput{
		{ID: "user_1", Email: "a@x.com"},
		{ID: "user_2", Email: "b@x.com", Attributes: map[string]any{"vip": true}},
	})

	// Create + send a campaign.
	tpl, _ := ne.Templates.Create(ctx, nextengage.TemplateInput{
		Name: "Promo", Subject: "Hi {{ email }}", Body: "<p>Sale!</p>",
	})
	camp, _ := ne.Campaigns.Create(ctx, nextengage.CampaignInput{
		Name: "Spring", TemplateID: tpl.ID, FromEmail: "news@acme.com",
	})
	_, _ = ne.Campaigns.Send(ctx, camp.ID)
}
```

The API key is scoped to a single project; the SDK resolves the project id
automatically from `/me` and caches it for the client's lifetime. A `Client` is
safe for concurrent use.

## Configuration

`New` accepts functional options:

| Option | Description |
| --- | --- |
| `WithAPIKey(key)` | Project API key (`ne_...`). Falls back to `NEXTENGAGE_API_KEY`. |
| `WithBaseURL(url)` | API base URL. Defaults to `https://api.nextengage.io`. Use `http://localhost:3001` for local dev. |
| `WithHTTPClient(hc)` | Custom `*http.Client` for timeouts, transport, proxies. |

Every request method takes a `context.Context` as its first argument for
cancellation and deadlines.

## Attribute operators

Use these as an attribute **value** to mutate it server-side atomically instead
of replacing it. Setting a value to `nil` deletes the key.

```go
ne.Contacts.Upsert(ctx, nextengage.ContactInput{
	ID: "user_42",
	Attributes: map[string]any{
		"visits": nextengage.Increment(1),         // atomic numeric add
		"tags":   nextengage.ArrayUnion("pro", "vip"), // dedup-add to array
		"beta":   nextengage.ArrayRemove("old"),    // remove from array
		"trial":  nil,                              // delete the key
	},
})
```

## Writes

`Contacts.Upsert`, `Track`, and `BulkImport` persist **synchronously** and return
the result directly — the saved `Contact`, `{ OK, ContactID }`, and
`{ Imported, NewCount }` respectively.

## Error handling

Any non-2xx response returns a `*nextengage.Error` carrying the HTTP status,
the server's message, and the raw body. Detect it with `errors.As`:

```go
_, err := ne.Contacts.Get(ctx, "missing")
var apiErr *nextengage.Error
if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
	// handle missing contact
}
```

## License

MIT
