package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	gh "github.com/crosbymichael/octokat"
	"github.com/crosbymichael/pulls"
	"os"
	"path"
	"time"
)

var (
	m          *pulls.Maintainer
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
	issues, err := m.GetIssues("open", c.String("assigned"))
	if err != nil {
		pulls.WriteError("Error getting issues: %s", err)
	}

	fmt.Printf("%c[2K\r", 27)
	pulls.DisplayIssues(c, issues, c.Bool("no-trunc"))
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

	app.Name = "issues"
	app.Usage = "Manage github issues"
	app.Version = "0.0.1"

	client := gh.NewClient()

	org, name, err := pulls.GetOriginUrl()
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
