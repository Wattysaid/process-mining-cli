package db

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func bigQueryClient(projectID string, credentialsPath string) (*bigquery.Client, error) {
	if projectID == "" {
		return nil, fmt.Errorf("bigquery project ID is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if credentialsPath != "" {
		return bigquery.NewClient(ctx, projectID, option.WithCredentialsFile(credentialsPath))
	}
	return bigquery.NewClient(ctx, projectID)
}

// TestBigQueryReadOnly validates connectivity by listing datasets.
func TestBigQueryReadOnly(projectID string, credentialsPath string) error {
	client, err := bigQueryClient(projectID, credentialsPath)
	if err != nil {
		return err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	it := client.Datasets(ctx)
	_, err = it.Next()
	if err == iterator.Done {
		return nil
	}
	return err
}

// ListSchemasBigQuery returns dataset IDs for the project.
func ListSchemasBigQuery(projectID string, credentialsPath string) ([]string, error) {
	client, err := bigQueryClient(projectID, credentialsPath)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out := []string{}
	it := client.Datasets(ctx)
	for {
		dataset, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		out = append(out, dataset.DatasetID)
	}
	return out, nil
}

// ListTablesBigQuery returns table IDs for a dataset.
func ListTablesBigQuery(projectID string, datasetID string, credentialsPath string) ([]string, error) {
	if datasetID == "" {
		return nil, fmt.Errorf("bigquery dataset ID is required")
	}
	client, err := bigQueryClient(projectID, credentialsPath)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out := []string{}
	it := client.Dataset(datasetID).Tables(ctx)
	for {
		table, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		out = append(out, table.TableID)
	}
	return out, nil
}

// ExtractBigQueryToCSV runs a query and writes the result to CSV.
func ExtractBigQueryToCSV(projectID string, credentialsPath string, query string, outputPath string) (int64, error) {
	if query == "" {
		return 0, fmt.Errorf("query is required")
	}
	if outputPath == "" {
		return 0, fmt.Errorf("output path is required")
	}
	client, err := bigQueryClient(projectID, credentialsPath)
	if err != nil {
		return 0, err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	job, err := client.Query(query).Run(ctx)
	if err != nil {
		return 0, err
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return 0, err
	}
	if err := status.Err(); err != nil {
		return 0, err
	}

	it, err := job.Read(ctx)
	if err != nil {
		return 0, err
	}
	schema := it.Schema
	headers := make([]string, len(schema))
	for i, field := range schema {
		headers[i] = field.Name
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writer.Write(headers); err != nil {
		return 0, err
	}

	var rowCount int64
	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return rowCount, err
		}
		record := make([]string, len(row))
		for i, value := range row {
			if value == nil {
				record[i] = ""
			} else {
				record[i] = fmt.Sprint(value)
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
	return rowCount, nil
}
