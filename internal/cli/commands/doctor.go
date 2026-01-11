package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/pm-assist/pm-assist/internal/config"
	"github.com/pm-assist/pm-assist/internal/db"
	"github.com/pm-assist/pm-assist/internal/policy"
	"github.com/spf13/cobra"
)

// NewDoctorCmd returns the doctor command.
func NewDoctorCmd(global *app.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check environment readiness",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(global.ConfigPath)
			if err != nil {
				return err
			}
			policies := policy.FromConfig(cfg)
			projectPath := global.ProjectPath
			if projectPath == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				projectPath = cwd
			}

			pythonPath, pythonErr := resolvePythonPath(projectPath)
			if pythonErr != nil {
				fmt.Println("[ERROR] Python not found")
			} else {
				fmt.Printf("[SUCCESS] Python found: %s\n", pythonPath)
				if err := checkPythonDeps(pythonPath); err != nil {
					fmt.Printf("[WARN] Python dependencies missing or failed to import: %v\n", err)
				} else {
					fmt.Println("[SUCCESS] Python dependencies imported successfully.")
				}
			}

			dotPath, dotErr := exec.LookPath("dot")
			if dotErr != nil {
				fmt.Println("[WARN] Graphviz not found (dot missing)")
			} else {
				fmt.Printf("[SUCCESS] Graphviz found: %s\n", dotPath)
			}

			if len(cfg.Connectors) == 0 {
				fmt.Println("[INFO] No connectors configured.")
			}

			for _, connector := range cfg.Connectors {
				switch connector.Type {
				case "file":
					if connector.File == nil {
						fmt.Printf("[WARN] File connector %s missing file config\n", connector.Name)
						continue
					}
					for _, path := range connector.File.Paths {
						if _, err := os.Stat(path); err != nil {
							fmt.Printf("[ERROR] File connector %s path not accessible: %s (%v)\n", connector.Name, path, err)
						} else {
							fmt.Printf("[SUCCESS] File connector %s path reachable: %s\n", connector.Name, path)
						}
					}
				case "database":
					if policies.OfflineOnly {
						fmt.Printf("[WARN] Offline-only policy enabled; skipping database check for %s.\n", connector.Name)
						continue
					}
					if connector.Database == nil || connector.Options == nil {
						fmt.Printf("[WARN] Database connector %s missing config\n", connector.Name)
						continue
					}
					driver := strings.ToLower(strings.TrimSpace(connector.Database.Driver))
					password := os.Getenv(connector.Options.CredentialEnv)
					if connector.Options.CredentialEnv == "" {
						fmt.Printf("[WARN] Database connector %s missing credential env var\n", connector.Name)
						continue
					}
					if password == "" {
						fmt.Printf("[WARN] Database connector %s credential env var not set: %s\n", connector.Name, connector.Options.CredentialEnv)
						continue
					}
					switch driver {
					case "postgres":
						dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s", connector.Database.Host, connector.Database.Port, connector.Database.DBName, connector.Database.User, password, connector.Database.SSLMode)
						if err := db.TestPostgresReadOnly(dsn); err != nil {
							fmt.Printf("[ERROR] Postgres connector %s failed: %v\n", connector.Name, err)
						} else {
							fmt.Printf("[SUCCESS] Postgres connector %s reachable.\n", connector.Name)
						}
					case "mysql":
						dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", connector.Database.User, password, connector.Database.Host, connector.Database.Port, connector.Database.DBName)
						if err := db.TestMySQLReadOnly(dsn); err != nil {
							fmt.Printf("[ERROR] MySQL connector %s failed: %v\n", connector.Name, err)
						} else {
							fmt.Printf("[SUCCESS] MySQL connector %s reachable.\n", connector.Name)
						}
					case "mssql":
						dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", connector.Database.User, password, connector.Database.Host, connector.Database.Port, connector.Database.DBName)
						if err := db.TestMSSQLReadOnly(dsn); err != nil {
							fmt.Printf("[ERROR] SQL Server connector %s failed: %v\n", connector.Name, err)
						} else {
							fmt.Printf("[SUCCESS] SQL Server connector %s reachable.\n", connector.Name)
						}
					case "snowflake":
						dsn, err := db.SnowflakeDSN(connector.Database.Host, connector.Database.User, password, connector.Database.DBName, connector.Database.Schema)
						if err != nil {
							fmt.Printf("[ERROR] Snowflake connector %s failed: %v\n", connector.Name, err)
							continue
						}
						if err := db.TestSnowflakeReadOnly(dsn); err != nil {
							fmt.Printf("[ERROR] Snowflake connector %s failed: %v\n", connector.Name, err)
						} else {
							fmt.Printf("[SUCCESS] Snowflake connector %s reachable.\n", connector.Name)
						}
					case "bigquery":
						if err := db.TestBigQueryReadOnly(connector.Database.DBName, password); err != nil {
							fmt.Printf("[ERROR] BigQuery connector %s failed: %v\n", connector.Name, err)
						} else {
							fmt.Printf("[SUCCESS] BigQuery connector %s reachable.\n", connector.Name)
						}
					default:
						fmt.Printf("[WARN] Connector %s uses unsupported driver %s.\n", connector.Name, driver)
					}
				default:
					fmt.Printf("[WARN] Unknown connector type %s for %s\n", connector.Type, connector.Name)
				}
			}

			if pythonErr != nil {
				return fmt.Errorf("environment check failed")
			}
			return nil
		},
		Example: "  pm-assist doctor",
	}
	return cmd
}

func resolvePythonPath(projectPath string) (string, error) {
	candidate := filepath.Join(projectPath, ".venv", "bin", "python")
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}
	pythonPath, err := exec.LookPath("python3")
	if err == nil {
		return pythonPath, nil
	}
	return exec.LookPath("python")
}

func checkPythonDeps(pythonPath string) error {
	cmd := exec.Command(pythonPath, "-c", "import pm4py, pandas, numpy, matplotlib, yaml, openpyxl, pyarrow")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}
