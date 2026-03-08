package appauth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestVerifySubjectConsistency(t *testing.T) {
	err := verifySubjectConsistency(
		map[string]any{"sub": "abc"},
		map[string]any{"sub": "abc"},
	)

	require.NoError(t, err)
}

func TestVerifySubjectConsistencyMissingUserInfoSubjectIsAccepted(t *testing.T) {
	err := verifySubjectConsistency(
		map[string]any{"sub": "abc"},
		map[string]any{},
	)

	require.NoError(t, err)
}

func TestVerifySubjectConsistencyMissingTokenSubjectIsAccepted(t *testing.T) {
	err := verifySubjectConsistency(
		map[string]any{},
		map[string]any{"sub": "abc"},
	)

	require.NoError(t, err)
}

func TestVerifySubjectConsistencyMismatch(t *testing.T) {
	err := verifySubjectConsistency(
		map[string]any{"sub": "abc"},
		map[string]any{"sub": "def"},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "subject mismatch")
}
