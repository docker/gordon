package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/aybabtme/color/brush"
	"github.com/codegangsta/cli"
	gh "github.com/crosbymichael/octokat"
	"github.com/dotcloud/gordon"
	"github.com/dotcloud/gordon/filters"
)

var (
	m *gordon.MaintainerManager
)

func displayAllPullRequests(c *cli.Context, state string, showAll bool) {
	filter := filters.GetPullRequestFilter(c)
	prs, err := filter(m.GetPullRequestsThatICareAbout(showAll, state, c.String("sort")))
	if err != nil {
		gordon.Fatalf("Error getting pull requests %s", err)
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
		gordon.Fatalf("Error getting pull requests: %s", err)
	}
	fmt.Printf("%v (#%d)\n", gordon.HumanDuration(time.Since(lru.UpdatedAt)), lru.Number)
}

func addComment(number, comment string) {
	cmt, err := m.AddComment(number, comment)
	if err != nil {
		gordon.Fatalf("%s", err)
	}
	gordon.DisplayCommentAdded(cmt)
}

func repositoryInfoCmd(c *cli.Context) {
	r, err := m.Repository()
	if err != nil {
		gordon.Fatalf("%s", err)
	}
	fmt.Printf("Name: %s\nForks: %d\nStars: %d\nIssues: %d\n", r.Name, r.Forks, r.Watchers, r.OpenIssues)
}

func mergeCmd(c *cli.Context) {
	if !c.Args().Present() {
		gordon.Fatalf("usage: merge ID")
	}
	number := c.Args()[0]
	merge, err := m.MergePullRequest(number, c.String("m"), c.Bool("force"))
	if err != nil {
		gordon.Fatalf("%s", err)
	}
	if merge.Merged {
		fmt.Printf("%s\n", brush.Green(merge.Message))
	} else {
		gordon.Fatalf("%s", err)
	}
}

func checkoutCmd(c *cli.Context) {
	if !c.Args().Present() {
		gordon.Fatalf("usage: checkout ID")
	}
	number := c.Args()[0]
	pr, _, err := m.GetPullRequest(number, false)
	if err != nil {
		gordon.Fatalf("%s", err)
	}
	if err := m.Checkout(pr); err != nil {
		gordon.Fatalf("%s", err)
	}
}

// Approve a pr by adding a LGTM to the comments
func approveCmd(c *cli.Context) {
	if !c.Args().Present() {
		gordon.Fatalf("usage: approve ID")
	}
	number := c.Args().First()
	if _, err := m.AddComment(number, "LGTM"); err != nil {
		gordon.Fatalf("%s", err)
	}
	fmt.Printf("Pull request %s approved\n", brush.Green(number))
}

// Show the patch in a PR
func showCmd(c *cli.Context) {
	if !c.Args().Present() {
		gordon.Fatalf("usage: show ID")
	}
	number := c.Args()[0]
	pr, _, err := m.GetPullRequest(number, false)
	if err != nil {
		gordon.Fatalf("%s", err)
	}
	patch, err := http.Get(pr.DiffURL)
	if err != nil {
		gordon.Fatalf("%s", err)
	}
	defer patch.Body.Close()

	if err := gordon.DisplayPatch(patch.Body); err != nil {
		gordon.Fatalf("%s", err)
	}
}

// Show contributors stats
func contributorsCmd(c *cli.Context) {
	contributors, err := m.GetContributors()
	if err != nil {
		gordon.Fatalf("%s", err)
	}
	gordon.DisplayContributors(c, contributors)
}

// Show the reviewers for this pull request
func reviewersCmd(c *cli.Context) {
	if !c.Args().Present() {
		gordon.Fatalf("usage: reviewers ID")
	}

	var (
		patch  io.Reader
		number = c.Args()[0]
	)

	if number == "-" {
		patch = os.Stdin
	} else {
		pr, _, err := m.GetPullRequest(number, false)
		if err != nil {
			gordon.Fatalf("%s", err)
		}

		resp, err := http.Get(pr.DiffURL)
		if err != nil {
			gordon.Fatalf("%s", err)
		}
		patch = resp.Body
		defer resp.Body.Close()
	}

	maintainers, err := gordon.GetMaintainersFromRepo(".")
	if err != nil {
		gordon.Fatalf("%s\n", err)
	}

	reviewers, err := gordon.ReviewPatch(patch, maintainers)
	if err != nil {
		gordon.Fatalf("%s", err)
	}
	gordon.DisplayReviewers(c, reviewers)
}

// This is the top level command for
// working with prs
func mainCmd(c *cli.Context) {
	if !c.Args().Present() {
		var (
			state   = c.String("state")
			showAll = true // default to true so that we get the fast path
		)
		switch {
		case c.Bool("no-merge"), c.Bool("lgtm"), c.Bool("new"), c.Bool("mine"):
			showAll = false
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
		gordon.Fatalf("%s", err)
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
			gordon.Fatalf("%s", err)
		}
	}
	if token != "" {
		config.Token = token
		if err := gordon.SaveConfig(*config); err != nil {
			gordon.Fatalf("%s", err)
		}
	}
	// Display token and user information
	if config, err := gordon.LoadConfig(); err == nil {
		if config.UserName != "" {
			fmt.Printf("Token: %s, UserName: %s\n", config.Token, config.UserName)
		} else {

			fmt.Printf("Token: %s\n", config.Token)
		}
	} else {
		fmt.Fprintf(os.Stderr, "No token registered\n")
		os.Exit(1)
	}
}

//Assign a pull request to the current user.
// If it's taken, show a message with the "--steal" optional flag.
//If the user doesn't have permissions, add a comment #volunteer
func takeCmd(c *cli.Context) {
	if !c.Args().Present() {
		gordon.Fatalf("usage: take ID")
	}
	number := c.Args()[0]
	pr, _, err := m.GetPullRequest(number, false)
	if err != nil {
		gordon.Fatalf("%s", err)
	}
	user, err := m.GetGithubUser()
	if err != nil {
		gordon.Fatalf("%s", err)
	}
	if pr.Assignee != nil && !c.Bool("overwrite") {
		gordon.Fatalf("Use --steal to steal the PR from %s", pr.Assignee.Login)
	}
	pr.Assignee = user
	patchedPR, err := m.PatchPullRequest(number, pr)
	if err != nil {
		gordon.Fatalf("%s", err)
	}
	if patchedPR.Assignee.Login != user.Login {
		m.AddComment(number, "#volunteer")
		fmt.Printf("No permission to assign. You '%s' was added as #volunteer.\n", user.Login)
	} else {
		fmt.Printf("Assigned PR %s to %s\n", brush.Green(number), patchedPR.Assignee.Login)
	}
}

func dropCmd(c *cli.Context) {
	if !c.Args().Present() {
		gordon.Fatalf("usage: drop ID")
	}
	number := c.Args()[0]
	pr, _, err := m.GetPullRequest(number, false)
	if err != nil {
		gordon.Fatalf("%s", err)
	}
	user, err := m.GetGithubUser()
	if err != nil {
		gordon.Fatalf("%s", err)
	}
	if pr.Assignee == nil || pr.Assignee.Login != user.Login {
		gordon.Fatalf("Can't drop %s: it's not yours.", number)
	}
	pr.Assignee = nil
	if _, err := m.PatchPullRequest(number, pr); err != nil {
		gordon.Fatalf("%s", err)
	}
	fmt.Printf("Unassigned PR %s\n", brush.Green(number))
}

func commentCmd(c *cli.Context) {
	if !c.Args().Present() {
		gordon.Fatalf("Please enter the issue's number")
	}
	number := c.Args()[0]
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}
	tmp, err := ioutil.TempFile("", "pulls-comment-")
	if err != nil {
		gordon.Fatalf("%s", err)
	}
	defer os.Remove(tmp.Name())
	cmd := exec.Command(editor, tmp.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		gordon.Fatalf("%v", err)
	}
	comment, err := ioutil.ReadAll(tmp)
	if err != nil {
		gordon.Fatalf("%v", err)
	}
	if _, err := m.AddComment(number, string(comment)); err != nil {
		gordon.Fatalf("%v", err)
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
		fmt.Fprintf(os.Stderr, "The current directory is not a valid git repository.\n")
		os.Exit(1)
	}
	t, err := gordon.NewMaintainerManager(client, org, name)
	if err != nil {
		gordon.Fatalf("%s", err)
	}
	m = t

	loadCommands(app)

	app.Run(os.Args)
}
