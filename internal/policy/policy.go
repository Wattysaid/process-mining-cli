package policy

import (
	"fmt"
	"strings"
)

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

// ValidateLLMProvider enforces LLM policy constraints.
func (p Policy) ValidateLLMProvider(provider string) error {
	key := strings.ToLower(strings.TrimSpace(provider))
	if key == "" {
		return nil
	}
	if p.LLMEnabled != nil && !*p.LLMEnabled && key != "none" {
		return fmt.Errorf("LLM provider %s is blocked by policy", provider)
	}
	if p.OfflineOnly && key != "none" && key != "ollama" {
		return fmt.Errorf("offline-only policy blocks external LLM provider %s", provider)
	}
	return nil
}
