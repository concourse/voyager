package main

import (
	"os"

	command "github.com/concourse/voyager/cli/command"
	"github.com/jessevdk/go-flags"
)

func main() {
	cmd := &command.MigrationCommand{}

	parser := flags.NewParser(cmd, flags.Default)

	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}
}
