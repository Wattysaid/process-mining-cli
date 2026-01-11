package profile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Profile represents a user profile stored in YAML.
type Profile struct {
	Name        string `yaml:"name"`
	Role        string `yaml:"role"`
	Aptitude    string `yaml:"aptitude"`
	PromptDepth string `yaml:"prompt_depth"`
}

// Save writes the profile to .profiles/<name>.yaml.
func Save(projectPath string, profile Profile) (string, error) {
	if profile.Name == "" {
		return "", errors.New("profile name is required")
	}
	profilesDir := filepath.Join(projectPath, ".profiles")
	if err := os.MkdirAll(profilesDir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(profilesDir, sanitizeFileName(profile.Name)+".yaml")
	content := fmt.Sprintf("name: %s\nrole: %s\naptitude: %s\npreferences:\n  prompt_depth: %s\n", profile.Name, profile.Role, profile.Aptitude, profile.PromptDepth)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// Load reads a profile from .profiles/<name>.yaml.
func Load(projectPath string, name string) (*Profile, error) {
	if name == "" {
		return nil, errors.New("profile name is required")
	}
	path := filepath.Join(projectPath, ".profiles", sanitizeFileName(name)+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Name        string `yaml:"name"`
		Role        string `yaml:"role"`
		Aptitude    string `yaml:"aptitude"`
		Preferences struct {
			PromptDepth string `yaml:"prompt_depth"`
		} `yaml:"preferences"`
	}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}
	return &Profile{
		Name:        parsed.Name,
		Role:        parsed.Role,
		Aptitude:    parsed.Aptitude,
		PromptDepth: parsed.Preferences.PromptDepth,
	}, nil
}

func sanitizeFileName(name string) string {
	out := make([]rune, 0, len(name))
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			out = append(out, r)
		case r >= 'A' && r <= 'Z':
			out = append(out, r)
		case r >= '0' && r <= '9':
			out = append(out, r)
		case r == '-' || r == '_':
			out = append(out, r)
		case r == ' ':
			out = append(out, '-')
		}
	}
	if len(out) == 0 {
		return "profile"
	}
	return string(out)
}
