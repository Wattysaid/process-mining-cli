package qa

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Results struct {
	RowCount           int                `json:"row_count"`
	MissingRates       map[string]float64 `json:"missing_rates"`
	DuplicateRate      float64            `json:"duplicate_rate"`
	OrderViolationRate float64            `json:"order_violation_rate"`
	TimestampParseRate float64            `json:"timestamp_parse_rate"`
	Warnings           []string           `json:"warnings"`
	BlockingIssues     []string           `json:"blocking_issues"`
	Thresholds         Thresholds         `json:"thresholds"`
}

type Thresholds struct {
	MissingValue float64 `json:"missing_value"`
	Duplicate    float64 `json:"duplicate"`
	OrderViol    float64 `json:"order_violation"`
	ParseFail    float64 `json:"parse_failure"`
}

type BacklogIssue struct {
	Severity string `json:"severity"`
	Issue    string `json:"issue"`
	Fix      string `json:"suggested_fix"`
}

func RunCSV(path string, caseCol string, activityCol string, timestampCol string, thresholds Thresholds) (Results, []BacklogIssue, error) {
	file, err := os.Open(path)
	if err != nil {
		return Results{}, nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return Results{}, nil, err
	}
	colIndex := make(map[string]int, len(header))
	for i, col := range header {
		colIndex[col] = i
	}

	missingCols := []string{}
	for _, col := range []string{caseCol, activityCol, timestampCol} {
		if _, ok := colIndex[col]; !ok {
			missingCols = append(missingCols, col)
		}
	}
	results := Results{
		MissingRates:   map[string]float64{},
		Warnings:       []string{},
		BlockingIssues: []string{},
		Thresholds:     thresholds,
	}
	if len(missingCols) > 0 {
		results.BlockingIssues = append(results.BlockingIssues, fmt.Sprintf("Missing required columns: %s", strings.Join(missingCols, ", ")))
		return results, []BacklogIssue{{Severity: "blocking", Issue: "Missing required columns", Fix: "Update column mapping or re-run ingest."}}, nil
	}

	missingCounts := map[string]int{caseCol: 0, activityCol: 0, timestampCol: 0}
	duplicateCount := 0
	rowCount := 0
	seen := map[string]struct{}{}

	caseLast := map[string]time.Time{}
	orderViolations := 0
	parsedTimestamps := 0
	parseFailures := 0

	for {
		record, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return results, nil, err
		}
		rowCount++
		caseVal := getValue(record, colIndex[caseCol])
		actVal := getValue(record, colIndex[activityCol])
		tsVal := getValue(record, colIndex[timestampCol])

		if strings.TrimSpace(caseVal) == "" {
			missingCounts[caseCol]++
		}
		if strings.TrimSpace(actVal) == "" {
			missingCounts[activityCol]++
		}
		if strings.TrimSpace(tsVal) == "" {
			missingCounts[timestampCol]++
		}

		key := caseVal + "|" + actVal + "|" + tsVal
		if _, ok := seen[key]; ok {
			duplicateCount++
		} else {
			seen[key] = struct{}{}
		}

		if strings.TrimSpace(tsVal) != "" {
			parsed, err := time.Parse(time.RFC3339, tsVal)
			if err != nil {
				parseFailures++
			} else {
				parsedTimestamps++
				if last, ok := caseLast[caseVal]; ok && parsed.Before(last) {
					orderViolations++
				}
				caseLast[caseVal] = parsed
			}
		}
	}

	results.RowCount = rowCount
	if rowCount == 0 {
		results.BlockingIssues = append(results.BlockingIssues, "No rows found in input log")
	}
	for col, count := range missingCounts {
		if rowCount > 0 {
			results.MissingRates[col] = float64(count) / float64(rowCount)
		}
	}
	if rowCount > 0 {
		results.DuplicateRate = float64(duplicateCount) / float64(rowCount)
	}
	if parsedTimestamps+parseFailures > 0 {
		results.TimestampParseRate = float64(parsedTimestamps) / float64(parsedTimestamps+parseFailures)
	}
	if rowCount > 0 {
		results.OrderViolationRate = float64(orderViolations) / float64(rowCount)
	}

	backlog := []BacklogIssue{}
	if results.RowCount == 0 {
		backlog = append(backlog, BacklogIssue{Severity: "blocking", Issue: "Empty dataset", Fix: "Check ingestion filters or source file."})
	}
	for col, rate := range results.MissingRates {
		if rate > thresholds.MissingValue {
			results.Warnings = append(results.Warnings, fmt.Sprintf("High missing rate for %s: %.2f", col, rate))
			backlog = append(backlog, BacklogIssue{Severity: "warning", Issue: fmt.Sprintf("Missing values above threshold in %s", col), Fix: "Review missingness strategy."})
		}
	}
	if results.DuplicateRate > thresholds.Duplicate {
		results.Warnings = append(results.Warnings, fmt.Sprintf("Duplicate rate above threshold: %.2f", results.DuplicateRate))
		backlog = append(backlog, BacklogIssue{Severity: "warning", Issue: "Duplicate rate above threshold", Fix: "Adjust dedupe keys or filter duplicates."})
	}
	if results.TimestampParseRate < 1.0-thresholds.ParseFail {
		results.Warnings = append(results.Warnings, fmt.Sprintf("Timestamp parse failures above threshold: %.2f", 1.0-results.TimestampParseRate))
		backlog = append(backlog, BacklogIssue{Severity: "warning", Issue: "Timestamp parse failures above threshold", Fix: "Specify timestamp format or clean source data."})
	}
	if results.OrderViolationRate > thresholds.OrderViol {
		results.Warnings = append(results.Warnings, fmt.Sprintf("Order violations above threshold: %.2f", results.OrderViolationRate))
		backlog = append(backlog, BacklogIssue{Severity: "warning", Issue: "Case order violations above threshold", Fix: "Sort by case and timestamp or review logging order."})
	}

	return results, backlog, nil
}

func WriteOutputs(outputDir string, results Results, backlog []BacklogIssue) error {
	qualityDir := filepath.Join(outputDir, "quality")
	if err := os.MkdirAll(qualityDir, 0o755); err != nil {
		return err
	}
	jsonPath := filepath.Join(qualityDir, "qa_results.json")
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonPath, jsonData, 0o644); err != nil {
		return err
	}

	mdPath := filepath.Join(qualityDir, "qa_summary.md")
	md := buildSummary(results, backlog)
	if err := os.WriteFile(mdPath, []byte(md), 0o644); err != nil {
		return err
	}

	csvPath := filepath.Join(qualityDir, "issues_backlog.csv")
	if err := writeCSV(csvPath, backlog); err != nil {
		return err
	}
	return nil
}

func buildSummary(results Results, backlog []BacklogIssue) string {
	lines := []string{
		"# QA Summary",
		"",
		fmt.Sprintf("- Rows: %d", results.RowCount),
		fmt.Sprintf("- Duplicate rate: %.2f", results.DuplicateRate),
		fmt.Sprintf("- Timestamp parse rate: %.2f", results.TimestampParseRate),
		fmt.Sprintf("- Order violation rate: %.2f", results.OrderViolationRate),
		"",
		"## Missing Rates",
	}
	for col, rate := range results.MissingRates {
		lines = append(lines, fmt.Sprintf("- %s: %.2f", col, rate))
	}
	lines = append(lines, "", "## Warnings")
	if len(results.Warnings) == 0 {
		lines = append(lines, "- None")
	} else {
		for _, warning := range results.Warnings {
			lines = append(lines, "- "+warning)
		}
	}
	lines = append(lines, "", "## Backlog")
	if len(backlog) == 0 {
		lines = append(lines, "- None")
	} else {
		for _, item := range backlog {
			lines = append(lines, fmt.Sprintf("- [%s] %s (Fix: %s)", item.Severity, item.Issue, item.Fix))
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

func writeCSV(path string, backlog []BacklogIssue) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	if err := writer.Write([]string{"severity", "issue", "suggested_fix"}); err != nil {
		return err
	}
	for _, item := range backlog {
		if err := writer.Write([]string{item.Severity, item.Issue, item.Fix}); err != nil {
			return err
		}
	}
	return nil
}

func getValue(record []string, idx int) string {
	if idx < 0 || idx >= len(record) {
		return ""
	}
	return record[idx]
}
