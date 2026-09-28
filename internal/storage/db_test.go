package storage

import (
	"testing"

	"fm-cli/internal/model"
)

// TestEmailDateRoundTrip checks a cached email comes back with the date it
// was saved with, not the driver's RFC 3339 form.
func TestEmailDateRoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	db, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	in := model.Email{ID: "e1", Subject: "Hi", Date: "2026-09-27 09:10", MailboxIDs: []string{"inbox"}}
	if err := db.SaveEmails([]model.Email{in}); err != nil {
		t.Fatal(err)
	}
	out, err := db.GetEmails("inbox", 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Date != in.Date {
		t.Fatalf("got %+v, want date %q", out, in.Date)
	}
}
