package main

import (
	"os"

	"github.com/homeport/freeze-calendar-resource/check"
	"github.com/homeport/freeze-calendar-resource/get"
	"github.com/homeport/freeze-calendar-resource/lint"
	"github.com/homeport/freeze-calendar-resource/put"
	"github.com/spf13/cobra"
)

func main() {
	if err := NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCommand = &cobra.Command{
	Use:   "freeze-calendar",
	Short: "Freeze Calendar Resource",
}

var lintCommand = cobra.Command{
	Use:   "lint",
	Short: "Checks syntax and semantics of a freeze calendar file",
	Args:  cobra.ExactArgs(1),
	RunE:  lint.RunE,
}

var checkCommand = cobra.Command{
	Use:   "check",
	Short: "Fetches the latest freeze calendar and emit its version",
	RunE: func(cmd *cobra.Command, args []string) error {
		return check.Check(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr())
	},
}

var getCommand = cobra.Command{
	Use:   "get",
	Short: "Fetches the latest version of the freeze calendar and, if within a freeze, fails or sleeps.",
	Long: `Fetches the latest version of the freeze calendar and

* If FUSE, the resource simply fails.
* If GATE, the resource sleeps while the current date and time are within a freeze window. This is re-tried every INTERVAL.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return get.Get(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(), args[0])
	},
}

var putCommand = cobra.Command{
	Use:   "put",
	Short: "no-op",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return put.Put(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(), args[0])
	},
}

func NewRootCommand() *cobra.Command {
	lintCommand.PersistentFlags().BoolVarP(&lint.Verbose, "verbose", "V", false, "verbose output")

	rootCommand.AddCommand(&lintCommand, &checkCommand, &getCommand, &putCommand)
	rootCommand.SilenceUsage = true

	return rootCommand
}
