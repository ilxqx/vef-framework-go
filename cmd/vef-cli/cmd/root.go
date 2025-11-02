package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ilxqx/vef-framework-go/cmd/vef-cli/cmd/build_info"
	"github.com/ilxqx/vef-framework-go/cmd/vef-cli/cmd/create"
	"github.com/ilxqx/vef-framework-go/cmd/vef-cli/cmd/model_schema"
)

var (
	Version string
	Date    string
)

var rootCmd = &cobra.Command{
	Use:   "vef-cli",
	Short: "VEF Framework CLI tool",
	Long:  `A command-line tool for VEF Framework to help with code generation and project setup.`,
}

// Execute runs the root command.
func Execute() {
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate(Banner + fmt.Sprintf("\nVersion: %s | Built: %s\n", Version, Date))

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)

		os.Exit(1)
	}
}

func init() {
	setupHelpColors(rootCmd)

	rootCmd.AddCommand(create.Command())
	rootCmd.AddCommand(build_info.Command())
	rootCmd.AddCommand(model_schema.Command())

	for _, cmd := range rootCmd.Commands() {
		setupHelpColors(cmd)
	}
}
