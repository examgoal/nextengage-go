package nextengage_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	nextengage "github.com/examgoal/nextengage-go"
)

func Example() {
	// apiKey falls back to the NEXTENGAGE_API_KEY env var when WithAPIKey is omitted.
	ne, err := nextengage.New(
		nextengage.WithAPIKey("ne_..."),
		nextengage.WithBaseURL("http://localhost:3001"), // omit for production
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	// Upsert a contact, keyed by YOUR own user id. Attributes MERGE; set a key to
	// nil to delete it.
	contact, err := ne.Contacts.Upsert(ctx, nextengage.ContactInput{
		ID:         "user_42",
		Email:      "ada@example.com",
		Attributes: map[string]any{"country": "IN", "plan": "pro"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(contact.ID)

	// Atomic attribute operators — no read-modify-write.
	_, _ = ne.Contacts.Upsert(ctx, nextengage.ContactInput{
		ID: "user_42",
		Attributes: map[string]any{
			"visits": nextengage.Increment(1),
			"tags":   nextengage.ArrayUnion("vip"),
		},
	})

	// Track a behavioral event (feeds dynamic segments).
	_, _ = ne.Contacts.Track(ctx, nextengage.Event{
		ID:         "user_42",
		Name:       "purchased",
		Properties: map[string]any{"amount": 49},
	})

	// Create and send a campaign.
	tpl, _ := ne.Templates.Create(ctx, nextengage.TemplateInput{
		Name: "Promo", Subject: "Hi {{ email }}", Body: "<p>Sale!</p>",
	})
	camp, _ := ne.Campaigns.Create(ctx, nextengage.CampaignInput{
		Name: "Spring", TemplateID: tpl.ID, FromEmail: "news@acme.com",
	})
	if _, err := ne.Campaigns.Send(ctx, camp.ID); err != nil {
		var apiErr *nextengage.Error
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusPaymentRequired {
			log.Println("plan limit reached")
		}
	}
}
