package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func listCmd() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all calendars on the server",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			calendars, err := clientFromContext(ctx).ListCalendars(ctx)
			if err != nil {
				return err
			}

			if len(calendars) == 0 {
				fmt.Println("No calendars found.")
				return nil
			}

			for _, cal := range calendars {
				fmt.Printf("Name: %s\n", cal.Name)
				fmt.Printf("Path: %s\n", cal.Path)
				if cal.Description != "" {
					fmt.Printf("Desc: %s\n", cal.Description)
				}
				fmt.Println("---")
			}

			return nil
		},
	}
}
