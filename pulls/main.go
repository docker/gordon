package main

import (
	"fmt"
	"github.com/aybabtme/color/brush"
	"github.com/codegangsta/cli"
	gh "github.com/crosbymichael/octokat"
	"github.com/crosbymichael/pulls"
	"os"
	"path"
	"strings"
	"time"
)

const (
	defaultTimeFormat = time.RFC822
)

var (
	m          *pulls.Maintainer
	configPath = path.Join(os.Getenv("HOME"), ".maintainercfg")
)

func listOpenPullsCmd(c *cli.Context) {
	var (
		pulls []*gh.PullRequest
		err   error
	)

	if c.Bool("no-merge") {
		pulls, err = m.GetNoMergePullRequests()
	} else {
		pulls, err = m.GetPullRequests("open")
	}
	if err != nil {
		writeError("Error getting pull requests %s\n", err)
	}
	displayPullRequests(pulls)
}

func listClosedPullsCmd(c *cli.Context) {
	pulls, err := m.GetPullRequests("closed")
	if err != nil {
		writeError("Error getting pull requests %s\n", err)
	}
	displayPullRequests(pulls)
}

func displayPullRequests(pulls []*gh.PullRequest) {
	w := newTabwriter()
	for _, p := range pulls {
		fmt.Fprintf(w, "%d\t%s\t%s\n", p.Number, truncate(p.Title), p.CreatedAt.Format(defaultTimeFormat))
	}

	if err := w.Flush(); err != nil {
		writeError("%s\n", err)
	}
}

func showPullRequestCmd(c *cli.Context) {
	if len(c.Args()) == 0 {
		writeError("%s\n", fmt.Errorf("Missing PR number"))
	}
	pr, err := m.GetPullRequest(c.Args()[0])
	if err != nil {
		writeError("%s\n", err)
	}
	displayPullRequest(pr)
}

func displayPullRequest(pr *gh.PullRequest) {
	fmt.Fprint(os.Stdout, brush.Green("Pull Request:"), "\n")
	fmt.Fprintf(os.Stdout, "No: %d\nTitle: %s\n\n", pr.Number, pr.Title)

	lines := strings.Split(pr.Body, "\n")
	for i, l := range lines {
		lines[i] = "\t" + l
	}
	fmt.Fprintf(os.Stdout, "Description:\n\n%s\n\n", strings.Join(lines, "\n"))
}

func repositoryInfoCmd(c *cli.Context) {
	r, err := m.Repository()
	if err != nil {
		writeError("%s\n", err)
	}
	fmt.Fprintf(os.Stdout, "Name: %s\nForks: %d\nStars: %d\nIssues: %d\n", r.Name, r.Forks, r.Watchers, r.OpenIssues)
}

func authCmd(c *cli.Context) {
	if token := c.String("add"); token != "" {
		if err := saveConfig(Config{token}); err != nil {
			writeError("%s\n", err)
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
	if len(c.Args()) == 0 {
		writeError("%s\n", fmt.Errorf("Missing PR number"))
	}
	number := c.Args()[0]
	pr, err := m.GetPullRequest(number)
	if err != nil {
		writeError("%s\n", err)
	}
	if c.Bool("add") {
		comment := c.Args()[1]
		cmt, err := m.AddComment(pr, comment)
		if err != nil {
			writeError("%s\n", err)
		}
		fmt.Fprintf(os.Stdout, "Comment added at %s\n", cmt.CreatedAt.Format(defaultTimeFormat))
		return
	} else {
		comments, err := m.GetComments(pr)
		if err != nil {
			writeError("%s\n", err)
		}
		fmt.Fprintln(os.Stdout, "Comments:\n")
		for _, c := range comments {
			fmt.Fprintf(os.Stdout, "@%s %s\n%s\n", brush.Red(c.User.Login), c.CreatedAt.Format(defaultTimeFormat), c.Body)
			fmt.Fprint(os.Stdout, "\n\n")
		}
	}
}

func loadCommands(app *cli.App) {
	app.Commands = []cli.Command{
		{
			Name:      "open",
			ShortName: "o",
			Usage:     "List all open pull requests for the current repository",
			Action:    listOpenPullsCmd,
			Flags: []cli.Flag{
				cli.BoolFlag{"no-merge", "display only prs that cannot be merged"},
			},
		},
		{
			Name:      "closed",
			ShortName: "c",
			Usage:     "List all closed pull requests for the current repository",
			Action:    listClosedPullsCmd,
		},
		{
			Name:      "show",
			ShortName: "s",
			Usage:     "Show the pull request based on the number",
			Action:    showPullRequestCmd,
		},
		{
			Name:      "repository",
			ShortName: "repo",
			Usage:     "List information about the current repository",
			Action:    repositoryInfoCmd,
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
			Name:      "comments",
			ShortName: "cmt",
			Usage:     "Show and manage comments for a pull reqeust",
			Action:    manageCommentsCmd,
			Flags: []cli.Flag{
				cli.BoolFlag{"add", "add a comment to the pull reqeust"},
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
