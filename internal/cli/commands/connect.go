package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/pm-assist/pm-assist/internal/cli/prompt"
	"github.com/pm-assist/pm-assist/internal/config"
	"github.com/pm-assist/pm-assist/internal/db"
	"github.com/pm-assist/pm-assist/internal/policy"
	"github.com/pm-assist/pm-assist/internal/preview"
	"github.com/spf13/cobra"
)

// NewConnectCmd returns the connect command.
func NewConnectCmd(global *app.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Register read-only data connectors",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectPath := global.ProjectPath
			if projectPath == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				projectPath = cwd
			}

			connectorType, err := prompt.AskChoice("Connector type", []string{"file", "database"}, "file", true)
			if err != nil {
				return err
			}

			cfg, err := config.Load(global.ConfigPath)
			if err != nil {
				return err
			}
			policies := policy.FromConfig(cfg)
			if cfg.Path == "" {
				cfg.Path = filepath.Join(projectPath, "pm-assist.yaml")
			}

			if !policies.AllowsConnector(connectorType) {
				return fmt.Errorf("connector type blocked by policy: %s", connectorType)
			}
			if policies.OfflineOnly && connectorType == "database" {
				return fmt.Errorf("database connectors are blocked in offline-only mode")
			}
			if connectorType == "file" {
				name, err := prompt.AskString("Connector name", "file-source", true)
				if err != nil {
					return err
				}
				pathList, err := prompt.AskString("File paths (comma-separated)", "", true)
				if err != nil {
					return err
				}
				format, err := prompt.AskChoice("Format", []string{"csv", "parquet"}, "csv", true)
				if err != nil {
					return err
				}
				delimiter := ""
				encoding := ""
				if format == "csv" {
					delimiter, err = prompt.AskString("CSV delimiter", ",", true)
					if err != nil {
						return err
					}
					encoding, err = prompt.AskString("CSV encoding", "utf-8", true)
					if err != nil {
						return err
					}
				}

				paths := splitCSV(pathList)
				for _, path := range paths {
					if _, err := os.Stat(path); err != nil {
						fmt.Printf("[WARN] Could not access %s: %v\n", path, err)
					}
				}
				previewNow, err := prompt.AskBool("Preview CSV headers and sample rows?", true)
				if err != nil {
					return err
				}
				if previewNow && format == "csv" {
					if strings.ToLower(encoding) != "utf-8" && encoding != "" {
						fmt.Printf("[WARN] Preview skipped for encoding %s (only utf-8 supported).\n", encoding)
					} else if len(paths) > 0 {
						countRows, err := prompt.AskBool("Count total rows? (may be slow)", false)
						if err != nil {
							return err
						}
						sample, err := preview.PreviewCSV(paths[0], delimiter, 5, countRows)
						if err != nil {
							fmt.Printf("[WARN] Preview failed: %v\n", err)
						} else {
							fmt.Println(preview.FormatSample(sample))
						}
					}
				}

				cfg.Connectors = append(cfg.Connectors, config.ConnectorSpec{
					Name: name,
					Type: "file",
					File: &config.FileConfig{
						Paths:     paths,
						Format:    format,
						Delimiter: delimiter,
						Encoding:  encoding,
					},
					Options: &config.ExtraConfig{ReadOnly: true},
				})
				if err := cfg.Save(); err != nil {
					return err
				}
				fmt.Println("[SUCCESS] File connector saved.")
				return nil
			}

			name, err := prompt.AskString("Connector name", "db-source", true)
			if err != nil {
				return err
			}
			driver, err := prompt.AskChoice("Database driver", []string{"postgres", "mysql", "mssql", "snowflake", "bigquery", "other"}, "postgres", true)
			if err != nil {
				return err
			}
			if !policies.AllowsConnector(driver) {
				return fmt.Errorf("connector driver blocked by policy: %s", driver)
			}
			host, err := prompt.AskString("Host", "", true)
			if err != nil {
				return err
			}
			portText, err := prompt.AskString("Port", "5432", true)
			if err != nil {
				return err
			}
			port, err := strconv.Atoi(portText)
			if err != nil {
				return fmt.Errorf("invalid port: %s", portText)
			}
			dbName, err := prompt.AskString("Database name", "", true)
			if err != nil {
				return err
			}
			schema, err := prompt.AskString("Schema (optional)", "", false)
			if err != nil {
				return err
			}
			user, err := prompt.AskString("Username (optional)", "", false)
			if err != nil {
				return err
			}
			sslMode, err := prompt.AskString("SSL mode (optional)", "", false)
			if err != nil {
				return err
			}
			credEnv, err := prompt.AskString("Credential env var name (e.g., DB_PASSWORD)", "", true)
			if err != nil {
				return err
			}

			fmt.Println("[INFO] Credentials are never stored in config. Set the env var before connecting.")
			fmt.Printf("[INFO] Using credential env var: %s\n", credEnv)

			testNow, err := prompt.AskBool("Test read-only connection now?", true)
			if err != nil {
				return err
			}
			if testNow {
				password := os.Getenv(credEnv)
				if password == "" {
					return fmt.Errorf("credential env var %s is not set", credEnv)
				}
				listCatalog, err := prompt.AskBool("List schemas and tables after validation?", false)
				if err != nil {
					return err
				}
				switch driver {
				case "postgres":
					dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s", host, port, dbName, user, password, sslMode)
					fmt.Println("[INFO] Testing Postgres read-only connection...")
					if err := db.TestPostgresReadOnly(dsn); err != nil {
						return fmt.Errorf("read-only test failed: %w", err)
					}
					fmt.Println("[SUCCESS] Read-only connection validated.")
					if listCatalog {
						if err := printCatalog(func() ([]string, error) { return db.ListSchemasPostgres(dsn) }, func(schema string) ([]string, error) { return db.ListTablesPostgres(dsn, schema) }); err != nil {
							return err
						}
					}
				case "mysql":
					dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, dbName)
					fmt.Println("[INFO] Testing MySQL read-only connection...")
					if err := db.TestMySQLReadOnly(dsn); err != nil {
						return fmt.Errorf("read-only test failed: %w", err)
					}
					fmt.Println("[SUCCESS] Read-only connection validated.")
					if listCatalog {
						if err := printCatalog(func() ([]string, error) { return db.ListSchemasMySQL(dsn) }, func(schema string) ([]string, error) { return db.ListTablesMySQL(dsn, schema) }); err != nil {
							return err
						}
					}
				case "mssql":
					dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", user, password, host, port, dbName)
					fmt.Println("[INFO] Testing SQL Server connection...")
					if err := db.TestMSSQLReadOnly(dsn); err != nil {
						return fmt.Errorf("read-only test failed: %w", err)
					}
					fmt.Println("[SUCCESS] Connection validated (read-only enforcement not guaranteed).")
					if listCatalog {
						if err := printCatalog(func() ([]string, error) { return db.ListSchemasMSSQL(dsn) }, func(schema string) ([]string, error) { return db.ListTablesMSSQL(dsn, schema) }); err != nil {
							return err
						}
					}
				default:
					return fmt.Errorf("driver %s is not supported for read-only validation yet", driver)
				}
			}

			cfg.Connectors = append(cfg.Connectors, config.ConnectorSpec{
				Name: name,
				Type: "database",
				Database: &config.DBConfig{
					Driver:  driver,
					Host:    host,
					Port:    port,
					DBName:  dbName,
					Schema:  schema,
					User:    user,
					SSLMode: sslMode,
				},
				Options: &config.ExtraConfig{ReadOnly: true, CredentialEnv: credEnv},
			})
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Println("[SUCCESS] Database connector saved.")
			return nil
		},
		Example: "  pm-assist connect",
	}
	return cmd
}

func splitCSV(input string) []string {
	parts := strings.Split(input, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func printCatalog(listSchemas func() ([]string, error), listTables func(schema string) ([]string, error)) error {
	schemas, err := listSchemas()
	if err != nil {
		return err
	}
	if len(schemas) == 0 {
		fmt.Println("[INFO] No schemas found.")
		return nil
	}
	fmt.Println("[INFO] Schemas:")
	for i, schema := range schemas {
		if i >= 10 {
			fmt.Println("[INFO] ...")
			break
		}
		fmt.Printf("  - %s\n", schema)
	}
	schema, err := prompt.AskString("Schema to list tables", schemas[0], true)
	if err != nil {
		return err
	}
	tables, err := listTables(schema)
	if err != nil {
		return err
	}
	fmt.Printf("[INFO] Tables in %s:\n", schema)
	for i, table := range tables {
		if i >= 20 {
			fmt.Println("[INFO] ...")
			break
		}
		fmt.Printf("  - %s\n", table)
	}
	return nil
}
