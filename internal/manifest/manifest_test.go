package manifest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestManifestLifecycle(t *testing.T) {
	dir := t.TempDir()
	runID := "run-123"
	manager, _, err := NewManager(runID, dir)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	inputPath := filepath.Join(dir, "input.txt")
	if err := os.WriteFile(inputPath, []byte("input"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}
	outputPath := filepath.Join(dir, "output.txt")
	if err := os.WriteFile(outputPath, []byte("output"), 0o644); err != nil {
		t.Fatalf("write output: %v", err)
	}
	if err := manager.StartStep("ingest"); err != nil {
		t.Fatalf("start step: %v", err)
	}
	if err := manager.AddInputs([]string{inputPath}); err != nil {
		t.Fatalf("add inputs: %v", err)
	}
	if err := manager.AddOutputs([]string{outputPath}); err != nil {
		t.Fatalf("add outputs: %v", err)
	}
	if err := manager.CompleteStep("ingest"); err != nil {
		t.Fatalf("complete step: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "run_manifest.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("unmarshal manifest: %v", err)
	}
	if manifest.RunID != runID {
		t.Fatalf("expected run id %s, got %s", runID, manifest.RunID)
	}
	if len(manifest.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(manifest.Steps))
	}
	if len(manifest.Inputs) != 1 || len(manifest.Outputs) != 1 {
		t.Fatalf("expected 1 input and 1 output entry")
	}
}
