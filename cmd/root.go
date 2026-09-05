// Package cmd is gdiff's command-line entry point.
package cmd

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"

	"github.com/JNSAPH/gdiff/internal/core"
	"github.com/JNSAPH/gdiff/internal/tui"
)

var (
	devFlag bool
	gitPath string
)

var rootCmd = &cobra.Command{
	Use:           "gdiff",
	Short:         "Browse a git repository's changes in the terminal",
	Args:          cobra.ExactArgs(0),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		core.SetupLogger(devFlag)

		if _, err := tea.NewProgram(tui.NewModel(gitPath)).Run(); err != nil {
			return fmt.Errorf("running program: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&devFlag, "dev", "d", false, "Run in development mode")
	rootCmd.Flags().StringVarP(&gitPath, "path", "p", ".", "Path to the git repository")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "gdiff:", err)
		os.Exit(1)
	}
}
