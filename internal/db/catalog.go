package db

import (
	"context"
	"database/sql"
	"time"
)

func ListSchemasPostgres(dsn string) ([]string, error) {
	return listSchemas("postgres", dsn, "SELECT schema_name FROM information_schema.schemata ORDER BY schema_name")
}

func ListTablesPostgres(dsn string, schema string) ([]string, error) {
	return listTables("postgres", dsn, "SELECT table_name FROM information_schema.tables WHERE table_schema=$1 ORDER BY table_name", schema)
}

func ListSchemasMySQL(dsn string) ([]string, error) {
	return listSchemas("mysql", dsn, "SELECT schema_name FROM information_schema.schemata ORDER BY schema_name")
}

func ListTablesMySQL(dsn string, schema string) ([]string, error) {
	return listTables("mysql", dsn, "SELECT table_name FROM information_schema.tables WHERE table_schema=? ORDER BY table_name", schema)
}

func ListSchemasMSSQL(dsn string) ([]string, error) {
	return listSchemas("sqlserver", dsn, "SELECT name FROM sys.schemas ORDER BY name")
}

func ListTablesMSSQL(dsn string, schema string) ([]string, error) {
	return listTables("sqlserver", dsn, "SELECT table_name FROM information_schema.tables WHERE table_schema=@p1 ORDER BY table_name", schema)
}

func listSchemas(driver string, dsn string, query string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

func listTables(driver string, dsn string, query string, schema string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, query, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}
