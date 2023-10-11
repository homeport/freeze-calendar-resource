package lint

import (
	"fmt"
	"os"

	"github.com/homeport/freeze-calendar-resource/freeze"
	"github.com/spf13/cobra"
)

var Verbose bool

func Run(cmd *cobra.Command, args []string) error {
	calendarFile, err := os.Open(args[0])

	if err != nil {
		return fmt.Errorf("unable to read calendar file from path %s: %w", args[0], err)
	}

	calendar, err := freeze.LoadCalendar(calendarFile)

	if err != nil {
		return err
	}

	if Verbose {
		cmd.Print("Calendar is valid ")
		switch len(calendar.Windows) {
		case 0:
			cmd.Print("but has no windows.")
		case 1:
			cmd.Print("and has one window:")
		default:
			cmd.Printf("and has %d windows:", len(calendar.Windows))
		}
		cmd.Println()

		for _, w := range calendar.Windows {
			cmd.Println(w)
		}
	}

	return err
}
