package appauth

import "context"

type (
	// User holds information about a user after successful authentication
	User struct {
		Sub    string         `json:"sub,omitempty"`
		Email  string         `json:"email,omitempty"`
		Name   string         `json:"name,omitempty"`
		Groups []string       `json:"groups,omitempty"`
		Roles  []string       `json:"roles,omitempty"` // merged realm+client roles (best effort)
		Raw    map[string]any `json:"raw,omitempty"`
	}
)

// UserFromContext extracts the User object from the request context
func UserFromContext(ctx context.Context) (*User, bool) {
	u, ok := ctx.Value(userKey).(*User)
	return u, ok
}
