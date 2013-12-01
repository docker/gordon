package main

import (
	"github.com/codegangsta/cli"
)

func loadCommands(app *cli.App) {
	// Add top level flags and commands
	app.Action = mainCmd

	// Filters modify what type of pr to display
	filters := []cli.Flag{
		cli.BoolFlag{"no-merge", "display only prs that cannot be merged"},
		cli.BoolFlag{"lgtm", "display the number of LGTM"},
		cli.BoolFlag{"closed", "display closed prs"},
		cli.BoolFlag{"new", "display prs opened in the last 24 hours"},
	}
	// Options modify how to display prs
	options := []cli.Flag{
		cli.BoolFlag{"no-trunc", "don't truncate pr name"},
		cli.StringFlag{"user", "", "display only prs from <user>"},
		cli.StringFlag{"comment", "", "add a comment to the pr"},
	}
	app.Flags = append(filters, options...)

	// Add subcommands
	app.Commands = []cli.Command{
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
		{
			Name:   "merge",
			Usage:  "Merge a pull request",
			Action: mergeCmd,
			Flags: []cli.Flag{
				cli.StringFlag{"m", "", "commit message for merge"},
				cli.BoolFlag{"force", "merge a pull request that has not been approved"},
			},
		},
		{
			Name:   "checkout",
			Usage:  "Checkout a pull request into your local repo",
			Action: checkoutCmd,
		},
		{
			Name:   "approve",
			Usage:  "Approve a pull request by adding LGTM to the comments",
			Action: approveCmd,
		},
	}
}
