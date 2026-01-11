package qa

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunCSVBasic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "log.csv")
	content := "case_id,activity,timestamp\n1,A,2024-01-01 10:00:00\n1,B,2024-01-01 11:00:00\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	thresholds := Thresholds{
		MissingValue: 0.1,
		Duplicate:    0.1,
		OrderViol:    0.1,
		ParseFail:    0.1,
	}
	results, backlog, err := RunCSV(path, "case_id", "activity", "timestamp", "", thresholds)
	if err != nil {
		t.Fatalf("run csv: %v", err)
	}
	if results.RowCount != 2 {
		t.Fatalf("expected 2 rows, got %d", results.RowCount)
	}
	if len(backlog) > 0 && results.RowCount == 0 {
		t.Fatalf("unexpected backlog for non-empty input")
	}
}
