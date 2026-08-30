package cmd

import (
	"context"

	"github.com/What-is-water93/caldav-cli/internal/caldav"
	"github.com/urfave/cli/v3"
)

func Root(version string) *cli.Command {
	return &cli.Command{
		Name:                  "caldav-cli",
		Usage:                 "CalDAV command-line client",
		Version:               version,
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "url", Usage: "CalDAV server URL", Sources: cli.EnvVars("CALDAV_URL"), Required: true},
			&cli.StringFlag{Name: "username", Usage: "CalDAV username", Sources: cli.EnvVars("CALDAV_USERNAME"), Required: true},
			&cli.StringFlag{Name: "password", Usage: "CalDAV password", Sources: cli.EnvVars("CALDAV_PASSWORD"), Required: true},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			client, err := caldav.NewClient(ctx, cmd.String("url"), cmd.String("username"), cmd.String("password"))
			if err != nil {
				return ctx, err
			}
			return context.WithValue(ctx, clientCtxKey{}, client), nil
		},
		Commands: []*cli.Command{
			listCmd(),
			uploadCmd(),
			eventsCmd(),
		},
	}
}
