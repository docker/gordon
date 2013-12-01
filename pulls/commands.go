package main

import (
	"github.com/codegangsta/cli"
)

func loadCommands(app *cli.App) {
	// Add top level flags and commands
	app.Action = mainCmd
	app.Flags = []cli.Flag{
		cli.BoolFlag{"no-trunc", "don't truncate pr name"},
		cli.BoolFlag{"no-merge", "display only prs that cannot be merged"},
		cli.BoolFlag{"lgtm", "display the number of LGTM"},
		cli.BoolFlag{"closed", "display closed prs"},
		cli.StringFlag{"user", "", "display only prs from <user>"},
		cli.StringFlag{"comment", "", "add a comment to the pr"},
	}

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
			},
		},
		{
			Name:   "checkout",
			Usage:  "Checkout a pull request into your local repo",
			Action: checkoutCmd,
		},
	}
}
