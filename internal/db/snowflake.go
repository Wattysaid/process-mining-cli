package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/snowflakedb/gosnowflake"
)

// SnowflakeDSN builds a DSN using the minimal required fields.
func SnowflakeDSN(account string, user string, password string, database string, schema string) (string, error) {
	if account == "" {
		return "", fmt.Errorf("snowflake account is required")
	}
	if user == "" {
		return "", fmt.Errorf("snowflake user is required")
	}
	cfg := &gosnowflake.Config{
		Account:  account,
		User:     user,
		Password: password,
		Database: database,
		Schema:   schema,
	}
	return gosnowflake.DSN(cfg)
}

// TestSnowflakeReadOnly validates connectivity by running a simple query.
func TestSnowflakeReadOnly(dsn string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := sql.Open("snowflake", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return err
	}
	var one int
	if err := db.QueryRowContext(ctx, "SELECT 1").Scan(&one); err != nil {
		return err
	}
	if one != 1 {
		return fmt.Errorf("unexpected response from read-only check")
	}
	return nil
}

// ListSchemasSnowflake returns schemas in the current database.
func ListSchemasSnowflake(dsn string) ([]string, error) {
	return listSchemas("snowflake", dsn, "SELECT schema_name FROM information_schema.schemata ORDER BY schema_name")
}

// ListTablesSnowflake returns tables for a schema.
func ListTablesSnowflake(dsn string, schema string) ([]string, error) {
	return listTables("snowflake", dsn, "SELECT table_name FROM information_schema.tables WHERE table_schema=? ORDER BY table_name", schema)
}
