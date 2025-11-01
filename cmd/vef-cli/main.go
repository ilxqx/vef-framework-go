package main

import (
	"github.com/ilxqx/vef-framework-go/cmd/vef-cli/cmd"
)

// Version information injected at build time via ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.Version = version
	cmd.Commit = commit
	cmd.Date = date
	cmd.Execute()
}
