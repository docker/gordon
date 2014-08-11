package main

import (
	"github.com/codegangsta/cli"
	"github.com/docker/gordon"
)

func loadCommands(app *cli.App) {
	app.Action = mainCmd

	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "assigned", Value: "", Usage: "display issues assigned to <user>. Use '*' for all assigned, or 'none' for all unassigned."},
		cli.StringFlag{Name: "remote", Value: gordon.GetDefaultGitRemote(), Usage: "git remote to treat as origin"},
		cli.StringFlag{Name: "milestone", Value: "", Usage: "display issues inside a particular <milestone>."},
		cli.BoolFlag{Name: "no-trunc", Usage: "do not truncate the issue name"},
		cli.IntFlag{Name: "votes", Value: -1, Usage: "display the number of votes '+1' filtered by the <number> specified."},
		cli.BoolFlag{Name: "vote", Usage: "add '+1' to an specific issue."},
		cli.BoolFlag{Name: "proposals", Usage: "Only show proposal issues"},
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
				cli.BoolFlag{Name: "overwrite", Usage: "overwrites a taken issue"},
			},
		},
		{
			Name:        "close",
			Usage:       "Close an issue",
			Description: "Provide the issue number for issue(s) to close for this repository",
			Action:      closeCmd,
			Flags:       []cli.Flag{},
		},
		{
			Name:   "search",
			Usage:  "Find issues by state and keyword.",
			Action: searchCmd,
			Flags: []cli.Flag{
				cli.StringFlag{Name: "author", Value: "", Usage: "Finds issues created by a certain user"},
				cli.StringFlag{Name: "assignee", Value: "", Usage: "Finds issues that are assigned to a certain user"},
				cli.StringFlag{Name: "mentions", Value: "", Usage: "Finds issues that mention a certain user"},
				cli.StringFlag{Name: "commenter", Value: "", Usage: "Finds issues that a certain user commented on"},
				cli.StringFlag{Name: "involves", Value: "", Usage: "Finds issues that were either created by a certain user, assigned to that user, mention that user, or were commented on by that user"},
				cli.StringFlag{Name: "labels", Value: "", Usage: "Filters issues based on their labels"},
				cli.StringFlag{Name: "state", Value: "", Usage: "Filter issues based on whether they’re open or closed"},
			},
		},
		{
			Name:   "auth",
			Usage:  "Add a github token for authentication",
			Action: authCmd,
			Flags: []cli.Flag{
				cli.StringFlag{Name: "add", Value: "", Usage: "add new token for authentication"},
			},
		},
	}
}
