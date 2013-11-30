package main

import (
	"github.com/codegangsta/cli"
)

func loadCommands(app *cli.App) {
	app.Commands = []cli.Command{
		{
			Name:   "open",
			Usage:  "List all open pull requests for the current repository",
			Action: listOpenPullsCmd,
			Flags: []cli.Flag{
				cli.BoolFlag{"no-trunc", "don't truncate pr name"},
				cli.BoolFlag{"no-merge", "display only prs that cannot be merged"},
				cli.BoolFlag{"lgtm", "display the number of LGTM"},
				cli.StringFlag{"user", "", "display only prs from <user>"},
			},
		},
		{
			Name:   "closed",
			Usage:  "List all closed pull requests for the current repository",
			Action: listClosedPullsCmd,
			Flags: []cli.Flag{
				cli.BoolFlag{"no-trunc", "don't truncate pr name"},
			},
		},
		{
			Name:   "show",
			Usage:  "Show the pull request based on the number",
			Action: showPullRequestCmd,
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
		{
			Name:   "comments",
			Usage:  "Show and manage comments for a pull request",
			Action: manageCommentsCmd,
			Flags: []cli.Flag{
				cli.BoolFlag{"add", "add a comment to the pull request"},
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
	}
}
