package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	ical "github.com/emersion/go-ical"
	"github.com/urfave/cli/v3"
	"golang.org/x/image/colornames"
)

const defaultColor = "indigo"

// validateColorName ensures name is a CSS3/SVG color name recognized by
// golang.org/x/image/colornames, and returns it normalized to lowercase.
func validateColorName(name string) (string, error) {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "" {
		return "", fmt.Errorf("color must not be empty")
	}
	if _, ok := colornames.Map[n]; !ok {
		return "", fmt.Errorf("unknown color name %q (expected a CSS3 color name like 'blue', 'red', 'cornflowerblue')", name)
	}
	return n, nil
}

func uploadCmd() *cli.Command {
	return &cli.Command{
		Name:  "upload",
		Usage: "Upload calendar events from an .ics file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "calendar",
				Aliases: []string{"c"},
				Usage:   "Path of the target calendar (auto-detected if only one exists)",
			},
			&cli.StringFlag{
				Name:     "file",
				Aliases:  []string{"f"},
				Usage:    "Path to the .ics file to upload",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "color",
				Usage: "Event color as a CSS3 color name.\nCheck https://pkg.go.dev/golang.org/x/image/colornames for valid values",
				Value: defaultColor,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			colorName, err := validateColorName(cmd.String("color"))
			if err != nil {
				return fmt.Errorf("invalid --color: %w", err)
			}

			f, err := os.Open(cmd.String("file"))
			if err != nil {
				return fmt.Errorf("opening file: %w", err)
			}
			defer f.Close()

			decoder := ical.NewDecoder(f)
			cal, err := decoder.Decode()
			if err != nil {
				return fmt.Errorf("decoding ICS file: %w", err)
			}

			client := clientFromContext(ctx)

			calendarPath, err := client.ResolveCalendar(ctx, cmd.String("calendar"))
			if err != nil {
				return err
			}

			count, err := client.UploadEvents(ctx, calendarPath, cal, colorName)
			if err != nil {
				return fmt.Errorf("after uploading %d event(s): %w", count, err)
			}

			fmt.Printf("Successfully uploaded %d event(s) with color %s.\n",
				count, colorizeName(colorName))
			return nil
		},
	}
}
