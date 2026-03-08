package appauth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractRoles(t *testing.T) {
	claims := map[string]any{
		"realm_access": map[string]any{
			"roles": []any{"admin", "offline_access", "admin"},
		},
		"resource_access": map[string]any{
			"myclient": map[string]any{
				"roles": []any{"api-read", "api-write", "api-read"},
			},
			"other-client": map[string]any{
				"roles": []any{"ignored"},
			},
		},
	}

	assert.Equal(t, []string{
		"admin",
		"myclient/api-read",
		"myclient/api-write",
		"offline_access",
	}, extractRoles(claims, "myclient"))
}

func TestExtractRolesWithoutClientID(t *testing.T) {
	claims := map[string]any{
		"realm_access": map[string]any{
			"roles": []any{"admin"},
		},
		"resource_access": map[string]any{
			"myclient": map[string]any{
				"roles": []any{"api-read"},
			},
		},
	}

	assert.Equal(t, []string{"admin"}, extractRoles(claims, ""))
}
