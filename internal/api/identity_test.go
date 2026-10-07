package api

import (
	"testing"

	"git.sr.ht/~rockorager/go-jmap/mail/identity"
)

func TestPreferredIdentity(t *testing.T) {
	ids := []*identity.Identity{
		{ID: "0", Email: "noah.hopping@fivesixhealthcare.com"},
		{ID: "3", Email: "noah@fivesixhealthcare.com"},
	}

	t.Run("match found", func(t *testing.T) {
		t.Setenv("FM_DEFAULT_FROM", "  Noah@FiveSixHealthcare.com ")
		got := PreferredIdentity(ids)
		if got == nil {
			t.Fatal("expected match, got nil")
		}
		if got.Email != "noah@fivesixhealthcare.com" {
			t.Fatalf("expected noah@, got %q", got.Email)
		}
	})

	t.Run("no match", func(t *testing.T) {
		t.Setenv("FM_DEFAULT_FROM", "nobody@example.com")
		if got := PreferredIdentity(ids); got != nil {
			t.Fatalf("expected nil, got %+v", got)
		}
	})

	t.Run("unset", func(t *testing.T) {
		t.Setenv("FM_DEFAULT_FROM", "")
		if got := PreferredIdentity(ids); got != nil {
			t.Fatalf("expected nil, got %+v", got)
		}
	})
}
