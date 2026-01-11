package policy

import "github.com/pm-assist/pm-assist/internal/config"

// FromConfig builds a policy from config values.
func FromConfig(cfg *config.Config) Policy {
	if cfg == nil {
		return Policy{}
	}
	return Policy{
		LLMEnabled:        cfg.Policy.LLMEnabled,
		OfflineOnly:       cfg.Policy.OfflineOnly,
		AllowedConnectors: cfg.Policy.AllowedConnectors,
		DeniedConnectors:  cfg.Policy.DeniedConnectors,
	}
}
