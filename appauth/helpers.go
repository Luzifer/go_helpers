package appauth

import (
	"slices"
	"strings"
)

func extractRoles(claims map[string]any, clientID string) []string {
	var out []string

	// realm_access.roles
	if ra, ok := claims["realm_access"].(map[string]any); ok {
		out = append(out, extractStringSlice(ra["roles"])...)
	}

	// resource_access[clientID].roles as explicit `clientID/role`
	if clientID != "" {
		if res, ok := claims["resource_access"].(map[string]any); ok {
			if c, ok := res[clientID].(map[string]any); ok {
				for _, roleName := range extractStringSlice(c["roles"]) {
					out = append(out, strings.Join([]string{clientID, roleName}, "/"))
				}
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
