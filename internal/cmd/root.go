package cmd

import "github.com/urfave/cli/v3"

func Root() *cli.Command {
	return &cli.Command{
		Name:  "caldav-cli",
		Usage: "CalDAV command-line client",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "url", Usage: "CalDAV server URL", Sources: cli.EnvVars("CALDAV_URL"), Required: true},
			&cli.StringFlag{Name: "username", Sources: cli.EnvVars("CALDAV_USERNAME"), Required: true},
			&cli.StringFlag{Name: "password", Sources: cli.EnvVars("CALDAV_PASSWORD"), Required: true},
		},
		Commands: []*cli.Command{
			listCmd(),
			// uploadCmd(),
		},
	}
}
