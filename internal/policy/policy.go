package policy

import "strings"

type Policy struct {
	LLMEnabled        *bool
	OfflineOnly       bool
	AllowedConnectors []string
	DeniedConnectors  []string
}

func normalize(items []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item))
		if key == "" {
			continue
		}
		out[key] = struct{}{}
	}
	return out
}

// AllowsConnector checks whether a connector type/driver is allowed.
func (p Policy) AllowsConnector(connector string) bool {
	key := strings.ToLower(strings.TrimSpace(connector))
	if key == "" {
		return true
	}
	deny := normalize(p.DeniedConnectors)
	if _, blocked := deny[key]; blocked {
		return false
	}
	allow := normalize(p.AllowedConnectors)
	if len(allow) == 0 {
		return true
	}
	_, ok := allow[key]
	return ok
}
