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
			if cfg.Path == "" {
				cfg.Path = filepath.Join(projectPath, "pm-assist.yaml")
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
				Options: &config.ExtraConfig{ReadOnly: true},
			})
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Println("[SUCCESS] Database connector saved.")
			fmt.Println("[INFO] Connection test will be available in a later release.")
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
