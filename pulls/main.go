package main

import (
	"fmt"
	"github.com/aybabtme/color/brush"
	"github.com/codegangsta/cli"
	gh "github.com/crosbymichael/octokat"
	"github.com/crosbymichael/pulls"
	"os"
)

var (
	m *pulls.Maintainer
)

func listOpenPullsCmd(c *cli.Context) {
	// FIXME: Pass a filter to the Getpullrequests method
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
	number := c.Args()[0]
	if comment := c.String("comment"); comment != "" {
		cmt, err := m.AddComment(number, comment)
		if err != nil {
			writeError("%s", err)
		}
		pulls.DisplayCommentAdded(cmt)
		os.Exit(0)
	}

	pr, comments, err := m.GetPullRequest(number, true)
	if err != nil {
		writeError("%s", err)
	}
	pulls.DisplayPullRequest(pr, comments)
}

func repositoryInfoCmd(c *cli.Context) {
	r, err := m.Repository()
	if err != nil {
		writeError("%s", err)
	}
	fmt.Fprintf(os.Stdout, "Name: %s\nForks: %d\nStars: %d\nIssues: %d\n", r.Name, r.Forks, r.Watchers, r.OpenIssues)
}

func mergeCmd(c *cli.Context) {
	number := c.Args()[0]
	merge, err := m.MergePullRequest(number, c.String("m"))
	if err != nil {
		writeError("%s", err)
	}
	if merge.Merged {
		fmt.Fprintf(os.Stdout, "%s\n", brush.Green(merge.Message))
	} else {
		writeError("%s", err)
	}
}

func checkoutCmd(c *cli.Context) {
	number := c.Args()[0]
	pr, _, err := m.GetPullRequest(number, false)
	if err != nil {
		writeError("%s", err)
	}
	if err := m.Checkout(pr); err != nil {
		writeError("%s", err)
	}
}

func main() {
	app := cli.NewApp()

	app.Name = "pulls"
	app.Usage = "Manage github pull requests for project maintainers"
	app.Version = "0.0.1"

	client := gh.NewClient()

	config := loadConfig()
	if config.Token != "" {
		client.WithToken(config.Token)
	}

	org, name, err := getOriginUrl()
	if err != nil {
		writeError("%s", err)
	}
	t, err := pulls.NewMaintainer(client, org, name)
	if err != nil {
		writeError("%s", err)
	}
	m = t

	loadCommands(app)

	app.Run(os.Args)
}
