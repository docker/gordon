package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	gh "github.com/crosbymichael/octokat"
	"github.com/crosbymichael/pulls"
	"github.com/crosbymichael/pulls/filters"
	"os"
	"path"
	"time"
)

var (
	m          *pulls.MaintainerManager
	configPath = path.Join(os.Getenv("HOME"), ".maintainercfg")
)

func alruCmd(c *cli.Context) {
	lru, err := m.GetFirstIssue("open", "updated")
	if err != nil {
		pulls.WriteError("Error getting issues: %s", err)
	}
	fmt.Printf("%v (#%d)\n", pulls.HumanDuration(time.Since(lru.UpdatedAt)), lru.Number)
}

func repositoryInfoCmd(c *cli.Context) {
	r, err := m.Repository()
	if err != nil {
		pulls.WriteError("%s", err)
	}
	fmt.Fprintf(os.Stdout, "Name: %s\nForks: %d\nStars: %d\nIssues: %d\n", r.Name, r.Forks, r.Watchers, r.OpenIssues)
}

//Take a specific issue. If it's taken, show a message with the overwrite optional flag
//If the user doesn't have permissions, add a comment #volunteer
func takeCmd(c *cli.Context) {
	if c.Args().Present() {
		number := c.Args()[0]
		issue, _, err := m.GetIssue(number, false)
		if err != nil {
			pulls.WriteError("%s", err)
		}
		user, err := m.GetGithubUser()
		if err != nil {
			pulls.WriteError("%s", err)
		}
		if issue.Assignee.Login != "" && !c.Bool("overwrite") {
			fmt.Fprintf(os.Stdout, "Use the flag --overwrite to take the issue from %s", issue.Assignee.Login)
			return
		}
		issue.Assignee = *user
		patchedIssue, err := m.PatchIssue(number, issue)
		if err != nil {
			pulls.WriteError("%s", err)
		}
		if patchedIssue.Assignee.Login != user.Login {
			m.AddComment(number, "#volunteer")
			fmt.Fprintf(os.Stdout, "No permission to assign. You '%s' was added as #volunteer.", user.Login)
		} else {
			fmt.Fprintf(os.Stdout, "The issue %s was assigned to %s", number, patchedIssue.Assignee.Login)
		}
	} else {
		fmt.Fprintf(os.Stdout, "Please enter the issue's number")
	}

}

func addComment(number, comment string) {
	cmt, err := m.AddComment(number, comment)
	if err != nil {
		pulls.WriteError("%s", err)
	}
	pulls.DisplayCommentAdded(cmt)
}

func mainCmd(c *cli.Context) {
	if !c.Args().Present() {
		filter := filters.GetIssueFilter(c)
		issues, err := filter(m.GetIssues("open", c.String("assigned")))
		if err != nil {
			pulls.WriteError("Error getting issues: %s", err)
		}

		fmt.Printf("%c[2K\r", 27)
		pulls.DisplayIssues(c, issues, c.Bool("no-trunc"))
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
	issue, comments, err := m.GetIssue(number, true)
	if err != nil {
		pulls.WriteError("%s", err)
	}
	pulls.DisplayIssue(issue, comments)
}

func authCmd(c *cli.Context) {
	config, err := pulls.LoadConfig()
	if err != nil {
		config = &pulls.Config{}
	}
	token := c.String("add")
	userName := c.String("user")
	if userName != "" {
		config.UserName = userName
		if err := pulls.SaveConfig(*config); err != nil {
			pulls.WriteError("%s", err)
		}
	}
	if token != "" {
		config.Token = token
		if err := pulls.SaveConfig(*config); err != nil {
			pulls.WriteError("%s", err)
		}
	}
	// Display token and user information
	if config, err := pulls.LoadConfig(); err == nil {
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

	app.Name = "issues"
	app.Usage = "Manage github issues"
	app.Version = "0.0.1"

	client := gh.NewClient()

	org, name, err := pulls.GetOriginUrl()
	if err != nil {
		panic(err)
	}
	t, err := pulls.NewMaintainerManager(client, org, name)
	if err != nil {
		panic(err)
	}
	m = t

	loadCommands(app)

	app.Run(os.Args)
}
