package main

import (
	"fmt"
	"github.com/aybabtme/color/brush"
	"github.com/codegangsta/cli"
	gh "github.com/crosbymichael/octokat"
	"github.com/crosbymichael/pulls"
	"os"
	"path"
)

var (
	m          *pulls.Maintainer
	configPath = path.Join(os.Getenv("HOME"), ".maintainercfg")
)

func listOpenPullsCmd(c *cli.Context) {
	prs, err := m.GetPullRequests("open")
	prs, err = m.FilterPullRequests(prs, c)
	if err != nil {
		writeError("Error getting pull requests %s", err)
	}
	fmt.Printf("%c[2K\r", 27)
	pulls.DisplayPullRequests(c, prs, c.Bool("no-trunc"))
}

func listClosedPullsCmd(c *cli.Context) {
	prs, err := m.GetPullRequests("closed")
	if err != nil {
		writeError("Error getting pull requests %s", err)
	}
	pulls.DisplayPullRequests(c, prs, c.Bool("no-trunc"))
}

func showPullRequestCmd(c *cli.Context) {
	pr, err := m.GetPullRequest(c.Args()[0])
	if err != nil {
		writeError("%s", err)
	}
	pulls.DisplayPullRequest(pr)
}

func repositoryInfoCmd(c *cli.Context) {
	r, err := m.Repository()
	if err != nil {
		writeError("%s", err)
	}
	fmt.Fprintf(os.Stdout, "Name: %s\nForks: %d\nStars: %d\nIssues: %d\n", r.Name, r.Forks, r.Watchers, r.OpenIssues)
}

func authCmd(c *cli.Context) {
	if token := c.String("add"); token != "" {
		if err := saveConfig(Config{token}); err != nil {
			writeError("%s", err)
		}
		return
	}
	// Display token and user information
	if config := loadConfig(); config.Token != "" {
		fmt.Fprintf(os.Stdout, "Token: %s\n", config.Token)
	} else {
		fmt.Fprintf(os.Stderr, "No token registered\n")
		os.Exit(1)
	}
}

func manageCommentsCmd(c *cli.Context) {
	number := c.Args()[0]
	if c.Bool("add") {
		comment := c.Args()[1]
		cmt, err := m.AddComment(number, comment)
		if err != nil {
			writeError("%s\n", err)
		}
		pulls.DisplayCommentAdded(cmt)
		return
	} else {
		comments, err := m.GetComments(number)
		if err != nil {
			writeError("%s\n", err)
		}
		pulls.DisplayComments(comments)
	}
}

func mergeCmd(c *cli.Context) {
	number := c.Args()[0]
	merge, err := m.MergePullRequest(number, c.String("m"))
	if err != nil {
		writeError("%s\n", err)
	}
	if merge.Merged {
		fmt.Fprintf(os.Stdout, "%s\n", brush.Green(merge.Message))
	} else {
		fmt.Fprintf(os.Stderr, "%s\n", brush.Red(merge.Message))
		os.Exit(1)
	}
}

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

func main() {
	app := cli.NewApp()

	app.Name = "pulls"
	app.Usage = "Manage github pull requets"
	app.Version = "0.0.1"

	client := gh.NewClient()

	config := loadConfig()
	if config.Token != "" {
		client.WithToken(config.Token)
	}

	org, name, err := getOriginUrl()
	if err != nil {
		panic(err)
	}
	t, err := pulls.NewMaintainer(client, org, name)
	if err != nil {
		panic(err)
	}
	m = t

	loadCommands(app)

	app.Run(os.Args)
}
