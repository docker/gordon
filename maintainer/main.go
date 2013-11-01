package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/crosbymichael/maintainer"
	gh "github.com/crosbymichael/octokat"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	defultTimeFormat = time.RFC822
)

var (
	m          *maintainer.Maintainer
	configPath = path.Join(os.Getenv("HOME"), ".maintainercfg")
)

func listIssues(c *cli.Context) {
	issues, err := m.GetIssues("open")
	if err != nil {
		writeError("Error getting issues %s", err)
	}

	w := newTabwriter()
	for _, i := range issues {
		fmt.Fprintf(w, "%d\t%s\t%s\n", i.Number, truncate(i.Title), i.CreatedAt.Format(defultTimeFormat))
	}
	if err := w.Flush(); err != nil {
		writeError("%s", err)
	}
}

func listPulls(c *cli.Context) {
	if len(c.Args()) > 0 {
		prId := c.Args()[0]
		pr, err := m.GetPullRequest(prId)
		if err != nil {
			writeError("%s", err)
		}
		displayPullRequest(pr)
		return
	}

	var (
		pulls []*gh.PullRequest
		err   error
	)

	if c.Bool("no-merge") {
		pulls, err = m.GetNoMergePullRequests()
	} else {
		pulls, err = m.GetPullRequests(c.String("state"))
	}
	if err != nil {
		writeError("Error getting pull reqeusts %s", err)
	}

	w := newTabwriter()
	for _, p := range pulls {
		fmt.Fprintf(w, "%d\t%s\t%s\n", p.Number, truncate(p.Title), p.CreatedAt.Format(defultTimeFormat))
	}

	if err := w.Flush(); err != nil {
		writeError("%s", err)
	}
}

func displayPullRequest(pr *gh.PullRequest) {
	fmt.Fprint(os.Stdout, "\033[1mPull Request:\033[0m\n")
	fmt.Fprintf(os.Stdout, "No: %d\nTitle: %s\n\n", pr.Number, pr.Title)

	lines := strings.Split(pr.Body, "\n")
	for i, l := range lines {
		lines[i] = "\t" + l
	}
	fmt.Fprintf(os.Stdout, "Description:\n\n%s\n\n", strings.Join(lines, "\n"))

	comments, err := m.GetComments(strconv.Itoa(pr.Number))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting comments: %s\n", err)
	} else {
		fmt.Fprintln(os.Stdout, "Comments:\n")
		for _, c := range comments {
			fmt.Fprintf(os.Stdout, "\033[1m@%s\033[0m %s\n%s\n", c.User.Login, c.CreatedAt.Format(defultTimeFormat), c.Body)
			fmt.Fprint(os.Stdout, "\n\n")
		}
	}
}

func repositoryInfo(c *cli.Context) {
	r, err := m.Repository()
	if err != nil {
		writeError("%s", err)
	}
	fmt.Fprintf(os.Stdout, "Name: %s\nForks: %d\nWatchers: %d\nIssues: %d\n", r.Name, r.Forks, r.Watchers, r.OpenIssues)
}

func addToken(c *cli.Context) {
	if len(c.Args()) == 0 {
		fmt.Fprintln(os.Stderr, "Missing token")
		os.Exit(1)
	}
	token := c.Args()[0]
	if err := saveConfig(Config{token}); err != nil {
		writeError("%s", err)
	}
}

func loadCommands(app *cli.App) {
	app.Commands = []cli.Command{
		{
			Name:      "issues",
			ShortName: "i",
			Usage:     "List all issues for the current repository",
			Action:    listIssues,
		},
		{
			Name:      "pulls",
			ShortName: "p",
			Usage:     "List all pull requests for the current repository",
			Action:    listPulls,
			Flags: []cli.Flag{
				cli.StringFlag{"state", "open", "state of the pull request (open, closed)"},
				cli.BoolFlag{"no-merge", "display only prs that cannot be merged"},
			},
		},
		{
			Name:      "repository",
			ShortName: "r",
			Usage:     "List information about the current repository",
			Action:    repositoryInfo,
		},
		{
			Name:   "add-token",
			Usage:  "Add a github token for authentication",
			Action: addToken,
		},
	}
}

func main() {
	app := cli.NewApp()

	app.Name = "maintainer"
	app.Usage = "Manage github issues and prs"
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
	t, err := maintainer.NewMaintainer(client, org, name)
	if err != nil {
		panic(err)
	}
	m = t

	loadCommands(app)

	app.Run(os.Args)
}
