package cmd

import (
	"context"
	"fmt"
	"time"

	ical "github.com/emersion/go-ical"
	"github.com/urfave/cli/v3"
)

// propText returns the string value of an iCalendar property, or "" if absent.
func propText(ev ical.Event, name string) string {
	if p := ev.Props.Get(name); p != nil {
		return p.Value
	}
	return ""
}

// propTexts returns all string values for a (possibly multi-valued) property.
func propTexts(ev ical.Event, name string) []string {
	props := ev.Props.Values(name)
	out := make([]string, 0, len(props))
	for _, p := range props {
		if p.Value != "" {
			out = append(out, p.Value)
		}
	}
	return out
}

// printRecurrence prints recurrence-related properties if the event repeats.
func printRecurrence(ev ical.Event) {
	rrules := propTexts(ev, ical.PropRecurrenceRule)  // RRULE
	rdates := propTexts(ev, ical.PropRecurrenceDates) // RDATE
	exdates := propTexts(ev, ical.PropExceptionDates) // EXDATE
	recurID := propText(ev, ical.PropRecurrenceID)    // RECURRENCE-ID

	// Nothing to print if the event isn't recurring and isn't an override.
	if len(rrules) == 0 && len(rdates) == 0 && len(exdates) == 0 && recurID == "" {
		return
	}

	for _, r := range rrules {
		fmt.Printf("RRule:    %s\n", r)
	}
	for _, r := range rdates {
		fmt.Printf("RDate:    %s\n", r)
	}
	for _, r := range exdates {
		fmt.Printf("ExDate:   %s\n", r)
	}
	if recurID != "" {
		fmt.Printf("RecurID:  %s\n", recurID)
	}
}

func eventsCmd() *cli.Command {
	return &cli.Command{
		Name:  "events",
		Usage: "List all events in a calendar",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "calendar",
				Aliases: []string{"c"},
				Usage:   "Path of the calendar (auto-detected if only one exists)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			client := clientFromContext(ctx)

			calendarPath, err := client.ResolveCalendar(ctx, cmd.String("calendar"))
			if err != nil {
				return err
			}

			events, err := client.ListEvents(ctx, calendarPath)
			if err != nil {
				return err
			}

			if len(events) == 0 {
				fmt.Println("No events found.")
				return nil
			}

			for _, ev := range events {
				fmt.Printf("Summary:  %s\n", propText(ev, ical.PropSummary))
				fmt.Printf("UID:      %s\n", propText(ev, ical.PropUID))

				if start, err := ev.DateTimeStart(time.Local); err == nil && !start.IsZero() {
					fmt.Printf("Start:    %s\n", start.Format("2006-01-02 15:04"))
				}
				if end, err := ev.DateTimeEnd(time.Local); err == nil && !end.IsZero() {
					fmt.Printf("End:      %s\n", end.Format("2006-01-02 15:04"))
				}
				if loc := propText(ev, ical.PropLocation); loc != "" {
					fmt.Printf("Location: %s\n", loc)
				}
				if desc := propText(ev, ical.PropDescription); desc != "" {
					fmt.Printf("Desc:     %s\n", desc)
				}
				if color := propText(ev, "COLOR"); color != "" {
					fmt.Printf("Color:    %s\n", colorizeName(color))
				}

				printRecurrence(ev)

				fmt.Println("---")
			}

			return nil
		},
	}
}
