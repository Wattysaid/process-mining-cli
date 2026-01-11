package preview

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

// CSVPreview contains headers and sample rows.
type CSVPreview struct {
	Headers []string
	Samples [][]string
	Rows    int
}

// PreviewCSV reads headers and up to sampleRows rows. If countAll is true, it counts all rows.
func PreviewCSV(path string, delimiter string, sampleRows int, countAll bool) (CSVPreview, error) {
	file, err := os.Open(path)
	if err != nil {
		return CSVPreview{}, err
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))
	reader.FieldsPerRecord = -1
	if delimiter != "" {
		runes := []rune(delimiter)
		if len(runes) > 0 {
			reader.Comma = runes[0]
		}
	}

	headers, err := reader.Read()
	if err != nil {
		return CSVPreview{}, err
	}

	samples := make([][]string, 0, sampleRows)
	rowCount := 0
	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return CSVPreview{}, err
		}
		rowCount++
		if len(samples) < sampleRows {
			samples = append(samples, record)
		}
		if !countAll && len(samples) >= sampleRows {
			break
		}
	}

	return CSVPreview{Headers: headers, Samples: samples, Rows: rowCount}, nil
}

// FormatSample renders a sample block for CLI output.
func FormatSample(preview CSVPreview) string {
	if len(preview.Headers) == 0 {
		return "[INFO] No headers found."
	}
	lines := []string{
		"[INFO] Columns:",
		"  " + strings.Join(preview.Headers, ", "),
	}
	if len(preview.Samples) > 0 {
		lines = append(lines, "[INFO] Sample rows:")
		for _, row := range preview.Samples {
			lines = append(lines, "  "+strings.Join(row, ", "))
		}
	}
	if preview.Rows > 0 {
		lines = append(lines, fmt.Sprintf("[INFO] Rows counted: %d", preview.Rows))
	}
	return strings.Join(lines, "\n")
}
