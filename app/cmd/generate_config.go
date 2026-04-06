package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const defaultConfigYAML = `# =============================================================================
# dashi configuration file
# =============================================================================
# Generated with: dashi config
#
# Configuration is loaded in this order (last wins):
#   1. Built-in defaults
#   2. .env file (optional)
#   3. This YAML config file (optional, default: ./config.yaml)
#   4. Environment variables (prefix: DASHI_)
#
# Environment variable format: DASHI_<SECTION>_<KEY>, e.g.:
#   DASHI_SERVER_PORT=8085
# =============================================================================

# -----------------------------------------------------------------------------
# Server — main application HTTP server
# -----------------------------------------------------------------------------
Server:
  # IP address to bind to. Empty string means listen on all interfaces.
  BindIp: ""
  # Port to listen on.
  Port: 8085

# -----------------------------------------------------------------------------
# Observability — metrics / health-check HTTP server
# -----------------------------------------------------------------------------
Observability:
  # Set to false to disable the observability server entirely.
  Enabled: false
  # IP address to bind to. Empty string means listen on all interfaces.
  BindIp: ""
  # Port to listen on.
  Port: 9090

# -----------------------------------------------------------------------------
# DataDir — directory for database, sessions, and other data
# -----------------------------------------------------------------------------
# Relative paths are resolved from the working directory.
DataDir: "./data"

# -----------------------------------------------------------------------------
# Env — runtime environment settings
# -----------------------------------------------------------------------------
Env:
  # Log level: "debug", "info", "warn", or "error".
  LogLevel: "info"
  # Production mode. When true, debug features are disabled.
  Production: false

# -----------------------------------------------------------------------------
# Auth — authentication settings
# -----------------------------------------------------------------------------
Auth:
  # Set to true to require login. When false, all operations use DefaultUser.
  Enabled: false
  # Username used for all operations when auth is disabled.
  DefaultUser: "default"
`

func generateConfigCmd() *cobra.Command {
	var outputFile = "./config.yaml"

	cmd := &cobra.Command{
		Use:   "config",
		Short: "generate a default configuration file",
		Long:  "generate a YAML configuration file with all default values and comments explaining each option",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := os.Stat(outputFile); err == nil {
				return fmt.Errorf("file %s already exists, not overwriting", outputFile)
			}
			if err := os.WriteFile(outputFile, []byte(defaultConfigYAML), 0600); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}
			fmt.Printf("Configuration written to %s\n", outputFile)
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "o", outputFile, "output file path")
	return cmd
}
