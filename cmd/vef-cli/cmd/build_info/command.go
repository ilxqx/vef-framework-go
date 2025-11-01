package build_info

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Command returns the generate-build-info cobra command.
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-build-info",
		Short: "Generate build information for the application",
		Long: `Generate build information including app version, build time, and commit hash.

This command creates a Go source file with build information variables that can be
overridden at build time using ldflags. The generated file includes:

  - AppVersion: git tag/version (or "dev" if no tags)
  - BuildTime: build timestamp (e.g., "2022-08-08 02:30:00")
  - GitCommit: git commit hash

Example usage in go:generate:
  // Using installed vef-cli
  //go:generate vef-cli generate-build-info -o internal/vef/build_info.go -p vef

  // Or using full GitHub path (no installation required)
  //go:generate go run github.com/ilxqx/vef-framework-go/cmd/vef-cli@latest generate-build-info -o internal/vef/build_info.go -p vef

`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			output, _ := cmd.Flags().GetString("output")
			pkg, _ := cmd.Flags().GetString("package")

			gray := color.New(color.FgHiBlack)
			green := color.New(color.FgGreen)
			cyan := color.New(color.FgCyan)

			_, _ = cyan.Println("Generating build info...")
			_, _ = gray.Print("  Output file: ")
			_, _ = fmt.Println(output)
			_, _ = gray.Print("  Package: ")
			_, _ = fmt.Println(pkg)

			if err := Generate(output, pkg); err != nil {
				return fmt.Errorf("failed to generate build info: %w", err)
			}

			_, _ = green.Println("✓ Successfully generated build info file")

			return nil
		},
	}

	cmd.Flags().StringP("output", "o", "build_info.go", "Output file path")
	cmd.Flags().StringP("package", "p", "main", "Package name")

	return cmd
}
