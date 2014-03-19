package main

import (
	"fmt"
	"github.com/aybabtme/color/brush"
	"github.com/codegangsta/cli"
	gh "github.com/crosbymichael/octokat"
	"github.com/dotcloud/gordon"
	"github.com/dotcloud/gordon/filters"
	"io"
	"net/http"
	"os"
	"time"
)

var (
	m *gordon.MaintainerManager
)

func displayAllPullRequests(c *cli.Context, state string, showAll bool) {
	filter := filters.GetPullRequestFilter(c)
	prs, err := filter(m.GetPullRequestsThatICareAbout(showAll, state))
	if err != nil {
		gordon.WriteError("Error getting pull requests %s", err)
	}

	fmt.Printf("%c[2K\r", 27)
	gordon.DisplayPullRequests(c, prs, c.Bool("no-trunc"))
}

func displayAllPullRequestFiles(c *cli.Context, number string) {
	prfs, err := m.GetPullRequestFiles(number)
	if err == nil {
		i := 1
		for _, p := range prfs {
			fmt.Printf("%d: filename %s additions %d deletions %d\n", i, p.FileName, p.Additions, p.Deletions)
			i++
		}
	}
}

func alruCmd(c *cli.Context) {
	lru, err := m.GetFirstPullRequest("open", "updated")
	if err != nil {
		gordon.WriteError("Error getting pull requests: %s", err)
	}
	fmt.Printf("%v (#%d)\n", gordon.HumanDuration(time.Since(lru.UpdatedAt)), lru.Number)
}

func addComment(number, comment string) {
	cmt, err := m.AddComment(number, comment)
	if err != nil {
		gordon.WriteError("%s", err)
	}
	gordon.DisplayCommentAdded(cmt)
}

func repositoryInfoCmd(c *cli.Context) {
	r, err := m.Repository()
	if err != nil {
		gordon.WriteError("%s", err)
	}
	fmt.Fprintf(os.Stdout, "Name: %s\nForks: %d\nStars: %d\nIssues: %d\n", r.Name, r.Forks, r.Watchers, r.OpenIssues)
}

func mergeCmd(c *cli.Context) {
	if !c.Args().Present() {
		fmt.Println("Please enter a pull request number")
		return
	}
	number := c.Args()[0]
	merge, err := m.MergePullRequest(number, c.String("m"), c.Bool("force"))
	if err != nil {
		gordon.WriteError("%s", err)
	}
	if merge.Merged {
		fmt.Fprintf(os.Stdout, "%s\n", brush.Green(merge.Message))
	} else {
		gordon.WriteError("%s", err)
	}
}

func checkoutCmd(c *cli.Context) {
	if !c.Args().Present() {
		fmt.Println("Please enter a pull request number")
		return
	}
	number := c.Args()[0]
	pr, _, err := m.GetPullRequest(number, false)
	if err != nil {
		gordon.WriteError("%s", err)
	}
	if err := m.Checkout(pr); err != nil {
		gordon.WriteError("%s", err)
	}
}

// Approve a pr by adding a LGTM to the comments
func approveCmd(c *cli.Context) {
	if !c.Args().Present() {
		fmt.Println("Please enter a pull request number")
		return
	}
	number := c.Args().First()
	if _, err := m.AddComment(number, "LGTM"); err != nil {
		gordon.WriteError("%s", err)
	}
	fmt.Fprintf(os.Stdout, "Pull request %s approved\n", brush.Green(number))
}

// Show the patch in a PR
func showCmd(c *cli.Context) {
	if !c.Args().Present() {
		fmt.Println("Please enter a pull request number")
		return
	}
	number := c.Args()[0]
	pr, _, err := m.GetPullRequest(number, false)
	if err != nil {
		gordon.WriteError("%s", err)
	}
	patch, err := http.Get(pr.DiffURL)
	if err != nil {
		gordon.WriteError("%s", err)
	}
	if _, err := io.Copy(os.Stdout, patch.Body); err != nil {
		gordon.WriteError("%s", err)
	}
}

// Show contributors stats
func contributorsCmd(c *cli.Context) {
	contributors, err := m.GetContributors()
	if err != nil {
		gordon.WriteError("%s", err)
	}
	gordon.DisplayContributors(c, contributors)
}

// Show the reviewers for this pull request
func reviewersCmd(c *cli.Context) {
	if !c.Args().Present() {
		fmt.Println("Please enter a pull request number")
		return
	}
	number := c.Args()[0]
	var patch io.Reader
	if number == "-" {
		patch = os.Stdin
	} else {
		pr, _, err := m.GetPullRequest(number, false)
		if err != nil {
			gordon.WriteError("%s", err)
		}
		resp, err := http.Get(pr.DiffURL)
		if err != nil {
			gordon.WriteError("%s", err)
		}
		patch = resp.Body
	}
	reviewers, err := gordon.ReviewPatch(patch, m.GetMaintainersDirMap())
	if err != nil {
		gordon.WriteError("%s", err)
	}
	gordon.DisplayReviewers(c, reviewers)
}

// This is the top level command for
// working with prs
func mainCmd(c *cli.Context) {
	if !c.Args().Present() {
		var (
			state   = "open"
			showAll = true // default to true so that we get the fast path
		)
		if c.Bool("closed") {
			state = "closed"
			showAll = false
		} else {
			if c.Bool("no-merge") || c.Bool("lgtm") || c.Bool("new") {
				showAll = false
			}
		}
		displayAllPullRequests(c, state, showAll)
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
		gordon.WriteError("%s", err)
	}
	gordon.DisplayPullRequest(pr, comments)
}

func authCmd(c *cli.Context) {
	config, err := gordon.LoadConfig()
	if err != nil {
		config = &gordon.Config{}
	}
	token := c.String("add")
	userName := c.String("user")
	if userName != "" {
		config.UserName = userName
		if err := gordon.SaveConfig(*config); err != nil {
			gordon.WriteError("%s", err)
		}
	}
	if token != "" {
		config.Token = token
		if err := gordon.SaveConfig(*config); err != nil {
			gordon.WriteError("%s", err)
		}
	}
	// Display token and user information
	if config, err := gordon.LoadConfig(); err == nil {
		if config.UserName != "" {
			fmt.Fprintf(os.Stdout, "Token: %s, UserName: %s\n", config.Token, config.UserName)
		} else {

			fmt.Fprintf(os.Stdout, "Token: %s\n", config.Token)
		}
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

	org, name, err := gordon.GetOriginUrl()
	if err != nil {
		gordon.WriteError("%s", err)
	}
	t, err := gordon.NewMaintainerManager(client, org, name)
	if err != nil {
		gordon.WriteError("%s", err)
	}
	m = t

	loadCommands(app)

	app.Run(os.Args)
}
