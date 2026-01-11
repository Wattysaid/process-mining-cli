package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const SchemaVersion = 1

type Manifest struct {
	SchemaVersion  int         `json:"schema_version"`
	RunID          string      `json:"run_id"`
	StartedAt      string      `json:"started_at"`
	CompletedAt    string      `json:"completed_at,omitempty"`
	Status         string      `json:"status"`
	ConfigSnapshot string      `json:"config_snapshot,omitempty"`
	Steps          []Step      `json:"steps,omitempty"`
	Inputs         []FileEntry `json:"inputs,omitempty"`
	Outputs        []FileEntry `json:"outputs,omitempty"`
}

type Step struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	StartedAt   string `json:"started_at,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
	Message     string `json:"message,omitempty"`
}

type FileEntry struct {
	Path       string `json:"path"`
	SizeBytes  int64  `json:"size_bytes"`
	SHA256     string `json:"sha256"`
	ModifiedAt string `json:"modified_at"`
}

type Manager struct {
	path    string
	baseDir string
	mu      sync.Mutex
}

func NewManager(runID string, outputDir string) (*Manager, *Manifest, error) {
	if runID == "" {
		return nil, nil, errors.New("run ID is required")
	}
	if outputDir == "" {
		return nil, nil, errors.New("output directory is required")
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, nil, err
	}
	manifestPath := filepath.Join(outputDir, "run_manifest.json")
	manager := &Manager{path: manifestPath, baseDir: outputDir}
	manifest, err := manager.loadOrCreate(runID)
	return manager, manifest, err
}

func (m *Manager) StartStep(name string) error {
	return m.updateStep(name, "started", "")
}

func (m *Manager) CompleteStep(name string) error {
	return m.updateStep(name, "completed", "")
}

func (m *Manager) FailStep(name string, message string) error {
	return m.updateStep(name, "failed", message)
}

func (m *Manager) SetStatus(status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	manifest, err := m.load()
	if err != nil {
		return err
	}
	manifest.Status = status
	if status == "completed" || status == "failed" {
		manifest.CompletedAt = time.Now().UTC().Format(time.RFC3339)
	}
	return m.save(manifest)
}

func (m *Manager) SetConfigSnapshot(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	manifest, err := m.load()
	if err != nil {
		return err
	}
	manifest.ConfigSnapshot = path
	return m.save(manifest)
}

func (m *Manager) AddInputs(paths []string) error {
	return m.addFiles(paths, true)
}

func (m *Manager) AddOutputs(paths []string) error {
	return m.addFiles(paths, false)
}

func (m *Manager) addFiles(paths []string, isInput bool) error {
	if len(paths) == 0 {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	manifest, err := m.load()
	if err != nil {
		return err
	}
	entries, err := buildFileEntries(paths, m.baseDir)
	if err != nil {
		return err
	}
	if isInput {
		manifest.Inputs = mergeEntries(manifest.Inputs, entries)
	} else {
		manifest.Outputs = mergeEntries(manifest.Outputs, entries)
	}
	return m.save(manifest)
}

func (m *Manager) updateStep(name string, status string, message string) error {
	if name == "" {
		return errors.New("step name is required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	manifest, err := m.load()
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	index := -1
	for i, step := range manifest.Steps {
		if step.Name == name {
			index = i
			break
		}
	}
	if index == -1 {
		step := Step{Name: name, Status: status}
		if status == "started" {
			step.StartedAt = now
		}
		if status == "completed" || status == "failed" {
			step.CompletedAt = now
		}
		if message != "" {
			step.Message = message
		}
		manifest.Steps = append(manifest.Steps, step)
		return m.save(manifest)
	}
	step := manifest.Steps[index]
	step.Status = status
	if step.StartedAt == "" && status == "started" {
		step.StartedAt = now
	}
	if status == "completed" || status == "failed" {
		step.CompletedAt = now
	}
	if message != "" {
		step.Message = message
	}
	manifest.Steps[index] = step
	return m.save(manifest)
}

func (m *Manager) loadOrCreate(runID string) (*Manifest, error) {
	if _, err := os.Stat(m.path); err == nil {
		return m.load()
	}
	manifest := &Manifest{
		SchemaVersion: SchemaVersion,
		RunID:         runID,
		StartedAt:     time.Now().UTC().Format(time.RFC3339),
		Status:        "running",
	}
	if err := m.save(manifest); err != nil {
		return nil, err
	}
	return manifest, nil
}

func (m *Manager) load() (*Manifest, error) {
	data, err := os.ReadFile(m.path)
	if err != nil {
		return nil, err
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}

func (m *Manager) save(manifest *Manifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.path, data, 0o644)
}

func buildFileEntries(paths []string, baseDir string) ([]FileEntry, error) {
	entries := []FileEntry{}
	seen := map[string]struct{}{}
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if info.IsDir() {
			err := filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}
				entry, err := fileEntry(p, baseDir)
				if err != nil {
					return err
				}
				if _, ok := seen[entry.Path]; ok {
					return nil
				}
				seen[entry.Path] = struct{}{}
				entries = append(entries, entry)
				return nil
			})
			if err != nil {
				return nil, err
			}
			continue
		}
		entry, err := fileEntry(path, baseDir)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[entry.Path]; ok {
			continue
		}
		seen[entry.Path] = struct{}{}
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})
	return entries, nil
}

func fileEntry(path string, baseDir string) (FileEntry, error) {
	info, err := os.Stat(path)
	if err != nil {
		return FileEntry{}, err
	}
	hash, err := hashFile(path)
	if err != nil {
		return FileEntry{}, err
	}
	entryPath := filepath.Clean(path)
	if baseDir != "" {
		rel, err := filepath.Rel(baseDir, path)
		if err == nil && !strings.HasPrefix(rel, "..") {
			entryPath = filepath.ToSlash(rel)
		}
	}
	return FileEntry{
		Path:       entryPath,
		SizeBytes:  info.Size(),
		SHA256:     hash,
		ModifiedAt: info.ModTime().UTC().Format(time.RFC3339),
	}, nil
}

func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func mergeEntries(existing []FileEntry, incoming []FileEntry) []FileEntry {
	out := make([]FileEntry, 0, len(existing)+len(incoming))
	seen := map[string]struct{}{}
	for _, entry := range existing {
		out = append(out, entry)
		seen[entry.Path] = struct{}{}
	}
	for _, entry := range incoming {
		if _, ok := seen[entry.Path]; ok {
			continue
		}
		out = append(out, entry)
		seen[entry.Path] = struct{}{}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Path < out[j].Path
	})
	return out
}
