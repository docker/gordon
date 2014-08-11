package main

import (
	"github.com/codegangsta/cli"
	"github.com/docker/gordon"
)

func loadCommands(app *cli.App) {
	// Add top level flags and commands
	app.Action = mainCmd

	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "remote", Value: gordon.GetDefaultGitRemote(), Usage: "git remote to treat as origin"},
	}

	// Filters modify what type of pr to display
	filters := []cli.Flag{
		cli.BoolFlag{Name: "no-merge", Usage: "display only prs that cannot be merged"},
		cli.BoolFlag{Name: "lgtm", Usage: "display the number of LGTM"},
		cli.StringFlag{Name: "state", Value: "open", Usage: "display prs based on their state"},
		cli.BoolFlag{Name: "new", Usage: "display prs opened in the last 24 hours"},
		cli.BoolFlag{Name: "mine", Usage: "display only PRs I care about based on the MAINTAINERS files"},
		cli.StringFlag{Name: "maintainer", Value: "", Usage: "display only PRs a maintainer cares about based on the MAINTAINERS files"},
		cli.StringFlag{Name: "sort", Value: "updated", Usage: "sort the prs by (created, updated, popularity, long-running)"},
		cli.StringFlag{Name: "assigned", Value: "", Usage: "display only prs assigned to a user"},
		cli.BoolFlag{Name: "unassigned", Usage: "display only unassigned prs"},
		cli.StringFlag{Name: "dir", Value: "", Usage: "display only prs that touch this dir"},
		cli.StringFlag{Name: "extension", Value: "", Usage: "display only prs that have files with this extension (no dot)"},
		cli.BoolFlag{Name: "cleanup", Usage: "display only cleanup prs"},
	}
	app.Flags = append(app.Flags, filters...)

	// Options modify how to display prs
	options := []cli.Flag{
		cli.BoolFlag{Name: "no-trunc", Usage: "don't truncate pr name"},
		cli.StringFlag{Name: "user", Value: "", Usage: "display only prs from <user>"},
		cli.StringFlag{Name: "comment", Value: "", Usage: "add a comment to the pr"},
	}
	app.Flags = append(app.Flags, options...)

	// Add subcommands
	app.Commands = []cli.Command{
		{
			Name:   "repo",
			Usage:  "List information about the current repository",
			Action: repositoryInfoCmd,
		},
		{
			Name:   "comment",
			Usage:  "Leave a comment on a pull request",
			Action: commentCmd,
		},
		{
			Name:   "comments",
			Usage:  "Show comments on a pull request",
			Action: commentsCmd,
		},
		{
			Name:   "auth",
			Usage:  "Add a github token for authentication",
			Action: authCmd,
			Flags: []cli.Flag{
				cli.StringFlag{Name: "add", Value: "", Usage: "add new token for authentication"},
				cli.StringFlag{Name: "user", Value: "", Usage: "add github user name"},
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
				cli.StringFlag{Name: "m", Value: "", Usage: "commit message for merge"},
				cli.BoolFlag{Name: "force", Usage: "merge a pull request that has not been approved"},
			},
		},
		{
			Name:   "close",
			Usage:  "Close a pull request without merging it",
			Action: closeCmd,
			Flags:  []cli.Flag{},
		},
		{
			Name:   "checkout",
			Usage:  "Checkout a pull request into your local repo",
			Action: checkoutCmd,
		},
		{
			Name:   "send",
			Usage:  "Send a new pull request, or overwrite an existing one",
			Action: sendCmd,
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
				cli.BoolFlag{Name: "steal", Usage: "steal the pull request from its current owner"},
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
				cli.BoolFlag{Name: "additions", Usage: "sort by additions"},
				cli.BoolFlag{Name: "deletions", Usage: "sort by deletions"},
				cli.BoolFlag{Name: "commits", Usage: "sort by commits"},
				cli.IntFlag{Name: "top", Value: 10, Usage: "top N contributors"},
			},
		},
	}
}
