package main

import (
	"fmt"
	"github.com/aybabtme/color/brush"
	"github.com/codegangsta/cli"
	gh "github.com/crosbymichael/octokat"
	"github.com/crosbymichael/pulls"
	"io"
	"net/http"
	"os"
	"time"
)

var (
	m *pulls.Maintainer
)

func displayAllPullRequests(c *cli.Context, state string) {
	filter := getFilter(c)
	prs, err := filter(m.GetPullRequests(state))
	if err != nil {
		pulls.WriteError("Error getting pull requests %s", err)
	}

	fmt.Printf("%c[2K\r", 27)
	pulls.DisplayPullRequests(c, prs, c.Bool("no-trunc"))
}

func alruCmd(c *cli.Context) {
	lru, err := m.GetFirstPullRequest("open", "updated")
	if err != nil {
		pulls.WriteError("Error getting pull requests: %s", err)
	}
	fmt.Printf("%v (#%d)\n", pulls.HumanDuration(time.Since(lru.UpdatedAt)), lru.Number)
}

func addComment(number, comment string) {
	cmt, err := m.AddComment(number, comment)
	if err != nil {
		pulls.WriteError("%s", err)
	}
	pulls.DisplayCommentAdded(cmt)
}

func repositoryInfoCmd(c *cli.Context) {
	r, err := m.Repository()
	if err != nil {
		pulls.WriteError("%s", err)
	}
	fmt.Fprintf(os.Stdout, "Name: %s\nForks: %d\nStars: %d\nIssues: %d\n", r.Name, r.Forks, r.Watchers, r.OpenIssues)
}

func mergeCmd(c *cli.Context) {
	number := c.Args()[0]
	merge, err := m.MergePullRequest(number, c.String("m"), c.Bool("force"))
	if err != nil {
		pulls.WriteError("%s", err)
	}
	if merge.Merged {
		fmt.Fprintf(os.Stdout, "%s\n", brush.Green(merge.Message))
	} else {
		pulls.WriteError("%s", err)
	}
}

func checkoutCmd(c *cli.Context) {
	number := c.Args()[0]
	pr, _, err := m.GetPullRequest(number, false)
	if err != nil {
		pulls.WriteError("%s", err)
	}
	if err := m.Checkout(pr); err != nil {
		pulls.WriteError("%s", err)
	}
}

// Approve a pr by adding a LGTM to the comments
func approveCmd(c *cli.Context) {
	number := c.Args().First()
	if _, err := m.AddComment(number, "LGTM"); err != nil {
		pulls.WriteError("%s", err)
	}
	fmt.Fprintf(os.Stdout, "Pull request %s approved\n", brush.Green(number))
}

// Show the patch in a PR
func showCmd(c *cli.Context) {
	number := c.Args()[0]
	pr, _, err := m.GetPullRequest(number, false)
	if err != nil {
		pulls.WriteError("%s", err)
	}
	patch, err := http.Get(pr.DiffURL)
	if err != nil {
		pulls.WriteError("%s", err)
	}
	if _, err := io.Copy(os.Stdout, patch.Body); err != nil {
		pulls.WriteError("%s", err)
	}
}

// Show the reviewers for this pull request
func reviewersCmd(c *cli.Context) {
	number := c.Args()[0]
	var patch io.Reader
	if number == "-" {
		patch = os.Stdin
	} else {
		pr, _, err := m.GetPullRequest(number, false)
		if err != nil {
			pulls.WriteError("%s", err)
		}
		resp, err := http.Get(pr.DiffURL)
		if err != nil {
			pulls.WriteError("%s", err)
		}
		patch = resp.Body
	}
	reviewers, err := ReviewPatch(patch)
	if err != nil {
		pulls.WriteError("%s", err)
	}
	for _, fileReviewers := range(reviewers) {
		for _, reviewer := range(fileReviewers) {
			fmt.Printf("%s\n", reviewer.Username)
		}
	}
}


// This is the top level command for
// working with prs
func mainCmd(c *cli.Context) {
	if !c.Args().Present() {
		state := "open"
		if c.Bool("closed") {
			state = "closed"
		}
		displayAllPullRequests(c, state)
		return
	}

	var (
		number  = c.Args().Get(0)
		comment = c.String("comment")
	)

	if comment != "" {
		addComment(number, comment)
		return
	}
	pr, comments, err := m.GetPullRequest(number, true)
	if err != nil {
		pulls.WriteError("%s", err)
	}
	pulls.DisplayPullRequest(pr, comments)
}

func authCmd(c *cli.Context) {
	if token := c.String("add"); token != "" {
		if err := pulls.SaveConfig(pulls.Config{token}); err != nil {
			pulls.WriteError("%s", err)
		}
		return
	}
	// Display token and user information
	if config, err := pulls.LoadConfig(); err == nil {
		fmt.Fprintf(os.Stdout, "Token: %s\n", config.Token)
	} else {
		fmt.Fprintf(os.Stderr, "No token registered\n")
		os.Exit(1)
	}
}

func main() {

	app := cli.NewApp()

	app.Name = "pulls"
	app.Usage = "Manage github pull requests for project maintainers"
	app.Version = "0.0.1"

	client := gh.NewClient()

	org, name, err := pulls.GetOriginUrl()
	if err != nil {
		pulls.WriteError("%s", err)
	}
	t, err := pulls.NewMaintainer(client, org, name)
	if err != nil {
		pulls.WriteError("%s", err)
	}
	m = t

	loadCommands(app)

	app.Run(os.Args)
}
