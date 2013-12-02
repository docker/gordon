package main

import (
	"github.com/codegangsta/cli"
)

func loadCommands(app *cli.App) {
	app.Commands = []cli.Command{
		{
			Name:   "list",
			Usage:  "List all open issues for the current repository",
			Action: listOpenIssuesCmd,
			Flags: []cli.Flag{
				cli.StringFlag{"assigned", "", "display issues assigned to <user>. Use '*' for all assigned, or 'none' for all unassigned."},
			},
		},
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

