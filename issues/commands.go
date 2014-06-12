package main

import (
	"github.com/codegangsta/cli"
)

func loadCommands(app *cli.App) {
	app.Action = mainCmd

	app.Flags = []cli.Flag{
		cli.StringFlag{"assigned", "", "display issues assigned to <user>. Use '*' for all assigned, or 'none' for all unassigned."},
		cli.StringFlag{"remote", "origin", "git remote to treat as origin"},
		cli.StringFlag{"milestone", "", "display issues inside a particular <milestone>."},
		cli.BoolFlag{"no-trunc", "do not truncate the issue name"},
		cli.IntFlag{"votes", -1, "display the number of votes '+1' filtered by the <number> specified."},
		cli.BoolFlag{"vote", "add '+1' to an specific issue."},
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
			Name:   "take",
			Usage:  "Assign an issue to your github account",
			Action: takeCmd,
			Flags: []cli.Flag{
				cli.BoolFlag{"overwrite", "overwrites a taken issue"},
			},
		},
		{
			Name:   "search",
			Usage:  "Find issues by state and keyword.",
			Action: searchCmd,
			Flags: []cli.Flag{
				cli.StringFlag{"author", "", "Finds issues created by a certain user"},
				cli.StringFlag{"assignee", "", "Finds issues that are assigned to a certain user"},
				cli.StringFlag{"mentions", "", "Finds issues that mention a certain user"},
				cli.StringFlag{"commenter", "", "Finds issues that a certain user commented on"},
				cli.StringFlag{"involves", "", "Finds issues that were either created by a certain user, assigned to that user, mention that user, or were commented on by that user"},
				cli.StringFlag{"labels", "", "Filters issues based on their labels"},
				cli.StringFlag{"state", "", "Filter issues based on whether they’re open or closed"},
			},
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
