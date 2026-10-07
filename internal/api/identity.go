package api

import (
	"os"
	"strings"

	"git.sr.ht/~rockorager/go-jmap/mail/identity"
)

// PreferredIdentity returns the sending identity named by the FM_DEFAULT_FROM
// environment variable, or nil when the variable is unset or matches no
// identity. Callers fall back to the first configured identity.
func PreferredIdentity(identities []*identity.Identity) *identity.Identity {
	want := strings.ToLower(strings.TrimSpace(os.Getenv("FM_DEFAULT_FROM")))
	if want == "" {
		return nil
	}
	for _, id := range identities {
		if id == nil {
			continue
		}
		if strings.ToLower(strings.TrimSpace(id.Email)) == want {
			return id
		}
	}
	return nil
}
