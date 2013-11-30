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
