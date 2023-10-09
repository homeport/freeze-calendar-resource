package main

import (
	"os"

	"github.com/homeport/freeze-calendar-resource/check"
	"github.com/homeport/freeze-calendar-resource/get"
	"github.com/spf13/cobra"
)

func main() {
	if err := NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

func NewRootCommand() *cobra.Command {
	var rootCommand = &cobra.Command{
		Use:   "freeze-calendar",
		Short: "Freeze Calendar Resource",
	}

	rootCommand.AddCommand(
		&cobra.Command{
			Use:   "check",
			Short: "Fetches the latest freeze calendar and emit its version",
			RunE:  check.Run,
		},
		&cobra.Command{
			Use:   "get",
			Short: "Fetches the latest version of the freeze calendar and, if within a freeze, fails or sleeps.",
			Long: `Fetches the latest version of the freeze calendar and:

* If FUSE, the resource simply fails.
* If GATE, the resource sleeps while the current date and time are within a freeze window. This is re-tried every INTERVAL.`,
			Args: cobra.ExactArgs(1),
			RunE: get.Run,
		},
		&cobra.Command{
			Use:   "put",
			Short: "not implemented",
			Run:   func(cmd *cobra.Command, args []string) { cmd.Println("no-op") },
		},
	)

	return rootCommand
}
