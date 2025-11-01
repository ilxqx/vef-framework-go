package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ilxqx/vef-framework-go/cmd/vef-cli/cmd/build_info"
	"github.com/ilxqx/vef-framework-go/cmd/vef-cli/cmd/create"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "vef-cli",
	Short: "VEF Framework CLI tool",
	Long:  `A command-line tool for VEF Framework to help with code generation and project setup.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate(Banner + fmt.Sprintf("\nVersion: %s | Commit: %s | Built: %s\n", Version, Commit, Date))

	setupHelpColors(rootCmd)

	rootCmd.AddCommand(create.Command())
	rootCmd.AddCommand(build_info.Command())

	for _, cmd := range rootCmd.Commands() {
		setupHelpColors(cmd)
	}
}
