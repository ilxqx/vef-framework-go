package main

import "github.com/ilxqx/vef-framework-go/cmd/vef-cli/cmd"

var (
	version = "0.0.1"
	date    = "2025-11-02 22:22:09"
)

func main() {
	cmd.Init(version, date)
	cmd.Execute()
}
