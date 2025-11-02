package cmd

import (
	"github.com/fatih/color"
)

// Banner is the VEF CLI ASCII art logo.
const Banner = `
██╗   ██╗███████╗███████╗     ██████╗██╗     ██╗
██║   ██║██╔════╝██╔════╝    ██╔════╝██║     ██║
██║   ██║█████╗  █████╗      ██║     ██║     ██║
╚██╗ ██╔╝██╔══╝  ██╔══╝      ██║     ██║     ██║
 ╚████╔╝ ███████╗██║         ╚██████╗███████╗██║
  ╚═══╝  ╚══════╝╚═╝          ╚═════╝╚══════╝╚═╝
`

// PrintBanner prints the banner and version info.
func PrintBanner() {
	cyan := color.New(color.FgCyan, color.Bold)
	_, _ = cyan.Print(Banner)

	_, _ = color.New(color.FgHiBlack).Printf("Version: %s | Built: %s\n\n",
		Version, Date)
}
