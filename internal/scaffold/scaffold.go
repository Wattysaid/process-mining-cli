package scaffold

import (
	"os"
	"path/filepath"
	"strings"
)

type Template struct {
	Name    string
	Folders []string
}

var StandardTemplate = Template{
	Name:    "standard",
	Folders: []string{"data", "outputs", "docs", ".profiles", ".business"},
}

// ApplyTemplate creates folders under projectPath.
func ApplyTemplate(projectPath string, template Template) error {
	for _, folder := range template.Folders {
		if err := os.MkdirAll(filepath.Join(projectPath, folder), 0o755); err != nil {
			return err
		}
	}
	return nil
}

// ParseCustomFolders parses comma-separated folder names.
func ParseCustomFolders(input string) []string {
	parts := strings.Split(input, ",")
	folders := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			folders = append(folders, trimmed)
		}
	}
	return folders
}

// EnsureGitignore adds standard entries to a .gitignore file if missing.
func EnsureGitignore(path string, entries []string) error {
	content := ""
	if data, err := os.ReadFile(path); err == nil {
		content = string(data)
	}
	needsNewline := content != "" && !strings.HasSuffix(content, "\n")
	lines := content
	for _, entry := range entries {
		if !strings.Contains(lines, entry) {
			if needsNewline {
				lines += "\n"
				needsNewline = false
			}
			lines += entry + "\n"
		}
	}
	return os.WriteFile(path, []byte(lines), 0o644)
}
