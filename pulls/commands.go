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
		cli.StringFlag{"state", "open", "display prs based on their state"},
		cli.BoolFlag{"new", "display prs opened in the last 24 hours"},
		cli.BoolFlag{"mine", "display only PRs I care about based on the MAINTAINERS files"},
		cli.StringFlag{"sort", "updated", "sort the prs by (created, updated, popularity, long-running)"},
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
				cli.StringFlag{"user", "", "add github user name"},
			},
		},
		{
			Name:   "alru",
			Usage:  "Show the Age of the Least Recently Updated pull request for this repo. Lower is better.",
			Action: alruCmd,
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
		{
			Name:   "take",
			Usage:  "Assign a pull request to your github account",
			Action: takeCmd,
			Flags: []cli.Flag{
				cli.BoolFlag{"steal", "steal the pull request from its current owner"},
			},
		},
		{
			Name:   "drop",
			Usage:  "Give up ownership of a pull request assigned to you",
			Action: dropCmd,
			Flags:  []cli.Flag{},
		},
		{
			Name:   "diff",
			Usage:  "Print the patch submitted by a pull request",
			Action: showCmd,
		},
		{
			Name:   "reviewers",
			Usage:  "Use the hierarchy of MAINTAINERS files to list who should review a pull request",
			Action: reviewersCmd,
		},
		{
			Name:   "contributors",
			Usage:  "Show the contributors list with additions, deletions, and commit counts. Default: sorted by Commits",
			Action: contributorsCmd,
			Flags: []cli.Flag{
				cli.BoolFlag{"additions", "sort by additions"},
				cli.BoolFlag{"deletions", "sort by deletions"},
				cli.BoolFlag{"commits", "sort by commits"},
				cli.IntFlag{"top", 10, "top N contributors"},
			},
		},
	}
}
