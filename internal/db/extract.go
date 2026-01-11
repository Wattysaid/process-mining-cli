package db

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

// ExtractQueryToCSV runs a query and writes results to CSV.
func ExtractQueryToCSV(driver string, dsn string, query string, outputPath string) (int64, error) {
	if query == "" {
		return 0, fmt.Errorf("query is required")
	}
	if outputPath == "" {
		return 0, fmt.Errorf("output path is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return 0, err
	}
	file, err := os.Create(outputPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writer.Write(columns); err != nil {
		return 0, err
	}

	values := make([]any, len(columns))
	valuePtrs := make([]any, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	var rowCount int64
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return rowCount, err
		}
		record := make([]string, len(columns))
		for i, value := range values {
			switch v := value.(type) {
			case nil:
				record[i] = ""
			case []byte:
				record[i] = string(v)
			default:
				record[i] = fmt.Sprint(v)
			}
		}
		if err := writer.Write(record); err != nil {
			return rowCount, err
		}
		rowCount++
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return rowCount, err
	}
	return rowCount, rows.Err()
}
