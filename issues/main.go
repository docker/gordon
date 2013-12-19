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

func mainCmd(c *cli.Context) {
	filter := filters.GetIssueFilter(c)
	issues, err := filter(m.GetIssues("open", c.String("assigned")))
	if err != nil {
		pulls.WriteError("Error getting issues: %s", err)
	}

	fmt.Printf("%c[2K\r", 27)
	pulls.DisplayIssues(c, issues, c.Bool("no-trunc"))
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
