package business

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Profile represents a business profile stored in YAML.
type Profile struct {
	Name     string
	Industry string
	Region   string
}

// Save writes the profile to .business/<name>.yaml.
func Save(projectPath string, profile Profile) (string, error) {
	if profile.Name == "" {
		return "", errors.New("business name is required")
	}
	businessDir := filepath.Join(projectPath, ".business")
	if err := os.MkdirAll(businessDir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(businessDir, sanitizeFileName(profile.Name)+".yaml")
	content := fmt.Sprintf("name: %s\nindustry: %s\nregion: %s\n", profile.Name, profile.Industry, profile.Region)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	return path, nil
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
		return "business"
	}
	return string(out)
}
