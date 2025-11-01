package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/ilxqx/vef-framework-go/constants"
)

func setupHelpColors(cmd *cobra.Command) {
	headerColor := color.New(color.FgCyan, color.Bold)
	commandColor := color.New(color.FgGreen)
	flagColor := color.New(color.FgGreen)

	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		if cmd.Parent() == nil {
			PrintBanner()
		}

		if cmd.Long != constants.Empty {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), cmd.Long)
			_, _ = fmt.Fprintln(cmd.OutOrStdout())
		} else if cmd.Short != constants.Empty {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), cmd.Short)
			_, _ = fmt.Fprintln(cmd.OutOrStdout())
		}

		_, _ = headerColor.Fprintln(cmd.OutOrStdout(), "Usage:")
		if cmd.Runnable() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s [flags]\n", cmd.CommandPath())
		}

		if cmd.HasAvailableSubCommands() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s [command]\n", cmd.CommandPath())
		}

		_, _ = fmt.Fprintln(cmd.OutOrStdout())

		if cmd.HasAvailableSubCommands() {
			_, _ = headerColor.Fprintln(cmd.OutOrStdout(), "Available Commands:")

			maxLen := 0
			for _, c := range cmd.Commands() {
				if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
					continue
				}

				if len(c.Name()) > maxLen {
					maxLen = len(c.Name())
				}
			}

			for _, c := range cmd.Commands() {
				if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
					continue
				}

				_, _ = commandColor.Fprintf(cmd.OutOrStdout(), "  %s", c.Name())
				spacing := maxLen - len(c.Name()) + 2
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%*s", spacing, constants.Space)
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), c.Short)
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout())
		}

		if cmd.HasAvailableLocalFlags() || cmd.HasAvailablePersistentFlags() {
			_, _ = headerColor.Fprintln(cmd.OutOrStdout(), "Flags:")

			if cmd.HasAvailableLocalFlags() {
				printFlags(cmd, cmd.LocalFlags(), flagColor)
			}

			if cmd.HasAvailableInheritedFlags() {
				_, _ = headerColor.Fprintln(cmd.OutOrStdout(), "\nGlobal Flags:")
				printFlags(cmd, cmd.InheritedFlags(), flagColor)
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout())
		}

		if cmd.HasAvailableSubCommands() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Use \"%s [command] --help\" for more information about a command.\n", cmd.CommandPath())
		}

		return nil
	})

	cmd.SetHelpFunc(func(*cobra.Command, []string) {
		_ = cmd.Usage()
	})
}

func printFlags(cmd *cobra.Command, flags any, flagColor *color.Color) {
	type FlagSet interface {
		FlagUsages() string
	}

	fs, ok := flags.(FlagSet)
	if !ok {
		return
	}

	usages := fs.FlagUsages()
	lines := strings.SplitSeq(usages, constants.Newline)

	for line := range lines {
		if line == constants.Empty {
			continue
		}

		trimmed := strings.TrimLeft(line, constants.Space)
		if trimmed == constants.Empty {
			continue
		}

		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), line)

			continue
		}

		flagPart := parts[0]
		descPart := strings.TrimLeft(parts[1], constants.Space)

		_, _ = flagColor.Fprint(cmd.OutOrStdout(), flagPart)
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", descPart)
	}
}
