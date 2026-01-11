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
	var (
		flagType        string
		flagName        string
		flagPaths       string
		flagFormat      string
		flagDelimiter   string
		flagEncoding    string
		flagSheet       string
		flagJSONLines   string
		flagZipMember   string
		flagPreview     string
		flagCountRows   string
		flagDriver      string
		flagHost        string
		flagPort        string
		flagDBName      string
		flagSchema      string
		flagUser        string
		flagSSLMode     string
		flagCredEnv     string
		flagTest        string
		flagListCatalog string
	)
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

			connectorType, err := resolveChoice(flagType, "Connector type", []string{"file", "database"}, "file", true)
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
				name, err := resolveString(flagName, "Connector name", "file-source", true)
				if err != nil {
					return err
				}
				pathList, err := resolveString(flagPaths, "File paths (comma-separated)", "", true)
				if err != nil {
					return err
				}
				format, err := resolveChoice(flagFormat, "Format", []string{"csv", "parquet", "xlsx", "json", "zip-csv", "xes"}, "csv", true)
				if err != nil {
					return err
				}
				delimiter := ""
				encoding := ""
				sheet := ""
				jsonLines := false
				zipMember := ""
				if format == "csv" || format == "zip-csv" {
					delimiter, err = resolveString(flagDelimiter, "CSV delimiter", ",", true)
					if err != nil {
						return err
					}
					encoding, err = resolveString(flagEncoding, "CSV encoding", "utf-8", true)
					if err != nil {
						return err
					}
					if format == "zip-csv" {
						zipMember, err = resolveString(flagZipMember, "ZIP member name (optional)", "", false)
						if err != nil {
							return err
						}
					}
				}
				if format == "xlsx" {
					sheet, err = resolveString(flagSheet, "Excel sheet name (optional)", "", false)
					if err != nil {
						return err
					}
				}
				if format == "json" {
					jsonLines, err = resolveBool(flagJSONLines, "JSON lines format?", false)
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
				previewNow, err := resolveBool(flagPreview, "Preview CSV headers and sample rows?", true)
				if err != nil {
					return err
				}
				if previewNow && (format == "csv" || format == "zip-csv") {
					if strings.ToLower(encoding) != "utf-8" && encoding != "" {
						fmt.Printf("[WARN] Preview skipped for encoding %s (only utf-8 supported).\n", encoding)
					} else if len(paths) > 0 {
						countRows, err := resolveBool(flagCountRows, "Count total rows? (may be slow)", false)
						if err != nil {
							return err
						}
						samplePath := paths[0]
						if format == "zip-csv" {
							fmt.Println("[WARN] CSV preview is skipped for zip archives.")
							countRows = false
						}
						if samplePath != "" && format != "zip-csv" {
							sample, err := preview.PreviewCSV(samplePath, delimiter, 5, countRows)
							if err != nil {
								fmt.Printf("[WARN] Preview failed: %v\n", err)
							} else {
								fmt.Println(preview.FormatSample(sample))
							}
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
						Sheet:     sheet,
						JSONLines: jsonLines,
						ZipMember: zipMember,
					},
					Options: &config.ExtraConfig{ReadOnly: true},
				})
				if err := cfg.Save(); err != nil {
					return err
				}
				fmt.Println("[SUCCESS] File connector saved.")
				return nil
			}

			name, err := resolveString(flagName, "Connector name", "db-source", true)
			if err != nil {
				return err
			}
			driver, err := resolveChoice(flagDriver, "Database driver", []string{"postgres", "mysql", "mssql", "snowflake", "bigquery", "other"}, "postgres", true)
			if err != nil {
				return err
			}
			if !policies.AllowsConnector(driver) {
				return fmt.Errorf("connector driver blocked by policy: %s", driver)
			}
			host, err := resolveString(flagHost, "Host", "", true)
			if err != nil {
				return err
			}
			portText, err := resolveString(flagPort, "Port", "5432", true)
			if err != nil {
				return err
			}
			port, err := strconv.Atoi(portText)
			if err != nil {
				return fmt.Errorf("invalid port: %s", portText)
			}
			dbName, err := resolveString(flagDBName, "Database name", "", true)
			if err != nil {
				return err
			}
			schema, err := resolveString(flagSchema, "Schema (optional)", "", false)
			if err != nil {
				return err
			}
			user, err := resolveString(flagUser, "Username (optional)", "", false)
			if err != nil {
				return err
			}
			sslMode, err := resolveString(flagSSLMode, "SSL mode (optional)", "", false)
			if err != nil {
				return err
			}
			credEnv, err := resolveString(flagCredEnv, "Credential env var name (e.g., DB_PASSWORD)", "", true)
			if err != nil {
				return err
			}

			fmt.Println("[INFO] Credentials are never stored in config. Set the env var before connecting.")
			fmt.Printf("[INFO] Using credential env var: %s\n", credEnv)

			testNow, err := resolveBool(flagTest, "Test read-only connection now?", true)
			if err != nil {
				return err
			}
			if testNow {
				password := os.Getenv(credEnv)
				if password == "" {
					return fmt.Errorf("credential env var %s is not set", credEnv)
				}
				listCatalog, err := resolveBool(flagListCatalog, "List schemas and tables after validation?", false)
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
				case "snowflake":
					dsn, err := db.SnowflakeDSN(host, user, password, dbName, schema)
					if err != nil {
						return err
					}
					fmt.Println("[INFO] Testing Snowflake connection...")
					if err := db.TestSnowflakeReadOnly(dsn); err != nil {
						return fmt.Errorf("read-only test failed: %w", err)
					}
					fmt.Println("[SUCCESS] Snowflake connection validated.")
					if listCatalog {
						if err := printCatalog(func() ([]string, error) { return db.ListSchemasSnowflake(dsn) }, func(schema string) ([]string, error) { return db.ListTablesSnowflake(dsn, schema) }); err != nil {
							return err
						}
					}
				case "bigquery":
					fmt.Println("[INFO] Testing BigQuery connection...")
					if err := db.TestBigQueryReadOnly(dbName, password); err != nil {
						return fmt.Errorf("read-only test failed: %w", err)
					}
					fmt.Println("[SUCCESS] BigQuery connection validated.")
					if listCatalog {
						if err := printCatalog(func() ([]string, error) { return db.ListSchemasBigQuery(dbName, password) }, func(schema string) ([]string, error) { return db.ListTablesBigQuery(dbName, schema, password) }); err != nil {
							return err
						}
					}
				default:
					fmt.Printf("[WARN] %s validation is not supported yet; skipping connection test.\n", driver)
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
	cmd.Flags().StringVar(&flagType, "type", "", "Connector type (file|database)")
	cmd.Flags().StringVar(&flagName, "name", "", "Connector name")
	cmd.Flags().StringVar(&flagPaths, "paths", "", "File paths (comma-separated)")
	cmd.Flags().StringVar(&flagFormat, "format", "", "File format (csv|parquet|xlsx|json|zip-csv|xes)")
	cmd.Flags().StringVar(&flagDelimiter, "delimiter", "", "CSV delimiter")
	cmd.Flags().StringVar(&flagEncoding, "encoding", "", "CSV encoding")
	cmd.Flags().StringVar(&flagSheet, "sheet", "", "Excel sheet name")
	cmd.Flags().StringVar(&flagJSONLines, "json-lines", "", "JSON lines format (true|false)")
	cmd.Flags().StringVar(&flagZipMember, "zip-member", "", "Zip member name")
	cmd.Flags().StringVar(&flagPreview, "preview", "", "Preview CSV headers and sample rows (true|false)")
	cmd.Flags().StringVar(&flagCountRows, "count-rows", "", "Count total rows when previewing (true|false)")
	cmd.Flags().StringVar(&flagDriver, "driver", "", "Database driver (postgres|mysql|mssql|snowflake|bigquery|other)")
	cmd.Flags().StringVar(&flagHost, "host", "", "Database host")
	cmd.Flags().StringVar(&flagPort, "port", "", "Database port")
	cmd.Flags().StringVar(&flagDBName, "database", "", "Database name")
	cmd.Flags().StringVar(&flagSchema, "schema", "", "Database schema")
	cmd.Flags().StringVar(&flagUser, "user", "", "Database user")
	cmd.Flags().StringVar(&flagSSLMode, "ssl-mode", "", "Database SSL mode")
	cmd.Flags().StringVar(&flagCredEnv, "credential-env", "", "Credential env var name")
	cmd.Flags().StringVar(&flagTest, "test", "", "Test read-only connection (true|false)")
	cmd.Flags().StringVar(&flagListCatalog, "list-catalog", "", "List schemas and tables after validation (true|false)")
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
