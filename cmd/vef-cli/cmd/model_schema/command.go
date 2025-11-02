package model_schema

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Command returns the generate-model-schema cobra command.
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-model-schema",
		Short: "Generate schema structures from Go models",
		Long: `Generate schema structures from Go models for ORM operations.

This command analyzes Go model structures that embed orm.BaseModel and generates
corresponding schema structures with type-safe column accessors. The generated
schemas help reduce hardcoded column name strings in ORM queries.

Features:
  - Processes single files or entire directories
  - Supports extend fields (bun:"extend")
  - Supports embed fields (bun:",embed:prefix_")
  - Extracts field labels from struct tags
  - Generates type-safe column accessors

Input/Output modes:
  - File → File: Process single model, generate single schema
  - Directory → Directory: Process all .go files, generate matching schemas
  - Directory → File: Not supported (will error)

Example usage:
  // Single file to file
  //go:generate vef-cli generate-model-schema -i models/user.go -o schemas/user.go -p schemas

  // Directory to directory
  //go:generate vef-cli generate-model-schema -i models -o schemas -p schemas

  // Using full GitHub path (no installation required)
  //go:generate go run github.com/ilxqx/vef-framework-go/cmd/vef-cli@latest generate-model-schema -i models -o schemas -p schemas
`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			input, _ := cmd.Flags().GetString("input")
			output, _ := cmd.Flags().GetString("output")
			pkg, _ := cmd.Flags().GetString("package")

			cyan := color.New(color.FgCyan)
			green := color.New(color.FgGreen)
			gray := color.New(color.FgHiBlack)
			red := color.New(color.FgRed)

			inputInfo, err := os.Stat(input)
			if err != nil {
				_, _ = red.Printf("✗ Input path does not exist: %s\n", input)
				return fmt.Errorf("input path does not exist: %w", err)
			}

			outputInfo, err := os.Stat(output)
			outputIsDir := err == nil && outputInfo.IsDir()
			inputIsDir := inputInfo.IsDir()

			if inputIsDir && !outputIsDir && err == nil {
				_, _ = red.Println("✗ When input is a directory, output must also be a directory")
				return fmt.Errorf("when input is a directory, output must also be a directory")
			}

			_, _ = cyan.Println("Generating model schemas...")
			_, _ = gray.Print("  Input: ")
			_, _ = fmt.Println(input)
			_, _ = gray.Print("  Output: ")
			_, _ = fmt.Println(output)
			_, _ = gray.Print("  Package: ")
			_, _ = fmt.Println(pkg)

			if inputIsDir {
				if err := GenerateDirectory(input, output, pkg); err != nil {
					return fmt.Errorf("failed to generate schemas: %w", err)
				}
			} else {
				if err := GenerateFile(input, output, pkg); err != nil {
					return fmt.Errorf("failed to generate schema: %w", err)
				}
			}

			_, _ = green.Println("✓ Successfully generated schema files")

			return nil
		},
	}

	cmd.Flags().StringP("input", "i", "", "Input model file or directory path")
	cmd.Flags().StringP("output", "o", "", "Output schema file or directory path")
	cmd.Flags().StringP("package", "p", "schemas", "Package name for generated schemas")

	_ = cmd.MarkFlagRequired("input")
	_ = cmd.MarkFlagRequired("output")

	return cmd
}
