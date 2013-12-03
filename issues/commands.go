package main

import (
	"github.com/codegangsta/cli"
)

func loadCommands(app *cli.App) {
	app.Action = mainCmd

	app.Flags = []cli.Flag{
		cli.StringFlag{"assigned", "", "display issues assigned to <user>. Use '*' for all assigned, or 'none' for all unassigned."},
		cli.BoolFlag{"no-trunc", "do not truncate the issue name"},
	}

	app.Commands = []cli.Command{
		{
			Name:   "alru",
			Usage:  "Show the Age of the Least Recently Updated issue for this repo. Lower is better.",
			Action: alruCmd,
		},
		{
			Name:   "repo",
			Usage:  "List information about the current repository",
			Action: repositoryInfoCmd,
		},
		{
			Name:   "auth",
			Usage:  "Add a github token for authentication",
			Action: authCmd,
			Flags: []cli.Flag{
				cli.StringFlag{"add", "", "add new token for authentication"},
			},
		},
	}
}
