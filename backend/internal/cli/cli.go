// Package cli handles command line operations
package cli

import (
	"path/filepath"

	"omar-kada/air-compose/internal/shell"
	"omar-kada/air-compose/internal/storage"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

// NewRootCmd creates a new command with default dependencies
func NewRootCmd(executor shell.Executor) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "air-compose",
		Short: "AirCompose CLI",
	}
	rootCmd.AddCommand(NewRunCommand(executor, func(params RunParams) (*gorm.DB, error) {
		return storage.NewGormDb(
			filepath.Join(params.GetDBDir(), "air-compose.db"),
			params.GetAddWritePerm(),
		)
	}))
	return rootCmd
}
