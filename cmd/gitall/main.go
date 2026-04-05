package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gitall/internal/status"

	"github.com/spf13/cobra"
)

var dirFlag string

var rootCmd = &cobra.Command{
	Use:   "gitall",
	Short: "gitall is a CLI utility to recursively walk directories with git projects",
	Long:  `gitall recursively walks subdirectories and collects git status from all repositories.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := dirFlag
		if dir == "" {
			dir = "."
		}

		absDir, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("invalid directory: %w", err)
		}

		return status.RunStatus(absDir)
	},
}

func main() {
	rootCmd.Flags().StringVar(&dirFlag, "dir", "", "starting directory (default: current directory)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
