package appauth

import (
	"slices"
	"strings"
)

func audContains(aud any, required string) bool {
	switch v := aud.(type) {
	case string:
		return v == required

	case []any:
		for _, it := range v {
			if s, ok := it.(string); ok && s == required {
				return true
			}
		}
	}

	return false
}

func bearerToken(hdr string) string {
	if hdr == "" {
		return ""
	}

	const p = "Bearer "
	if !strings.HasPrefix(hdr, p) {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(hdr, p))
}

func extractRoles(claims map[string]any, clientID string) []string {
	var out []string

	// realm_access.roles
	if ra, ok := claims["realm_access"].(map[string]any); ok {
		out = append(out, extractStringSlice(ra["roles"])...)
	}

	// resource_access[clientID].roles
	if clientID != "" {
		if res, ok := claims["resource_access"].(map[string]any); ok {
			if c, ok := res[clientID].(map[string]any); ok {
				out = append(out, extractStringSlice(c["roles"])...)
			}
		}
	}

	slices.Sort(out)
	return slices.Compact(out)
}

func extractStringSlice(v any) []string {
	switch x := v.(type) {
	case []string:
		return x

	case []any:
		out := make([]string, 0, len(x))
		for _, it := range x {
			if s, ok := it.(string); ok {
				out = append(out, s)
			}
		}
		return out

	default:
		return nil
	}
}

func str(v any) string {
	s, _ := v.(string)
	return s
}
