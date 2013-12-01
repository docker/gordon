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

func listOpenIssuesCmd(c *cli.Context) {
	issues, err := m.GetIssues("open", c.String("assigned"))
	if err != nil {
		writeError("Error getting issues: %s", err)
	}
	pulls.DisplayIssues(c, issues, c.Bool("no-trunc"))
}

func alruCmd(c *cli.Context) {
	lru, err := m.GetFirstIssue("open", "updated")
	if err != nil {
		writeError("Error getting issues: %s", err)
	}
	fmt.Printf("%v (#%d)\n", pulls.HumanDuration(time.Since(lru.UpdatedAt)), lru.Number)
}

func repositoryInfoCmd(c *cli.Context) {
	r, err := m.Repository()
	if err != nil {
		writeError("%s", err)
	}
	fmt.Fprintf(os.Stdout, "Name: %s\nForks: %d\nStars: %d\nIssues: %d\n", r.Name, r.Forks, r.Watchers, r.OpenIssues)
}

func main() {
	app := cli.NewApp()

	app.Name = "issues"
	app.Usage = "Manage github issues"
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
