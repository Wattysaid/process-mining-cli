package paths

import (
	"errors"
	"os"
	"path/filepath"
)

const skillsEnvVar = "PM_ASSIST_SKILLS_DIR"
const wheelsEnvVar = "PM_ASSIST_WHEELS_DIR"

// SkillsRoot resolves the skills directory, preferring env override, then project, then bundled assets.
func SkillsRoot(projectPath string) (string, error) {
	if override := os.Getenv(skillsEnvVar); override != "" {
		if exists(override) {
			return override, nil
		}
		return "", errors.New("skills path override not found")
	}

	if projectPath != "" {
		candidate := filepath.Join(projectPath, ".codex", "skills", "cli-tool-skills")
		if exists(candidate) {
			return candidate, nil
		}
	}

	exe, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "resources", "cli-tool-skills")
		if exists(candidate) {
			return candidate, nil
		}
	}

	return "", errors.New("cli-tool-skills not found; set PM_ASSIST_SKILLS_DIR")
}

func SkillPath(skillsRoot string, parts ...string) string {
	segments := append([]string{skillsRoot}, parts...)
	return filepath.Join(segments...)
}

func exists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// WheelsRoot resolves the bundled wheels directory for offline installs.
func WheelsRoot(projectPath string) (string, error) {
	if override := os.Getenv(wheelsEnvVar); override != "" {
		if exists(override) {
			return override, nil
		}
		return "", errors.New("wheels path override not found")
	}

	if projectPath != "" {
		candidate := filepath.Join(projectPath, "resources", "wheels")
		if exists(candidate) {
			return candidate, nil
		}
	}

	exe, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "resources", "wheels")
		if exists(candidate) {
			return candidate, nil
		}
	}

	return "", errors.New("bundled wheels not found; set PM_ASSIST_WHEELS_DIR")
}
