// Package nextengage is the official Go SDK for the NextEngage marketing
// platform: programmatic access to a single project using a project API key.
//
//	ne, err := nextengage.New(nextengage.WithAPIKey("ne_..."))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	ctx := context.Background()
//
//	// `ID` is YOUR own user id and IS the contact's key — same id, same contact.
//	_, err = ne.Contacts.Upsert(ctx, nextengage.ContactInput{
//	    ID:         "user_42",
//	    Email:      "ada@example.com",
//	    Attributes: map[string]any{"plan": "pro"},
//	})
//
//	// Atomic attribute operators (no read-modify-write):
//	_, err = ne.Contacts.Upsert(ctx, nextengage.ContactInput{
//	    ID: "user_42",
//	    Attributes: map[string]any{
//	        "visits": nextengage.Increment(1),
//	        "tags":   nextengage.ArrayUnion("vip"),
//	    },
//	})
//
//	_, err = ne.Contacts.Track(ctx, nextengage.Event{ID: "user_42", Name: "purchased"})
//
// The API key is scoped to one project; the project id is resolved automatically
// from the /me endpoint and cached for the lifetime of the client.
package nextengage
