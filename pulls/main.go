package main

import (
	"fmt"
	"github.com/aybabtme/color/brush"
	"github.com/codegangsta/cli"
	gh "github.com/crosbymichael/octokat"
	"github.com/crosbymichael/pulls"
	"github.com/crosbymichael/pulls/term"
	"github.com/nsf/termbox-go"
	"os"
	"path"
	"strconv"
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

func displayInteractiveCmd(c *cli.Context) {
	var err error
	if err = term.Init(); err != nil {
		writeError("%s\n", err)
	}

	screen, err := term.NewScreen(termbox.ColorGreen, termbox.ColorDefault)
	if err != nil {
		writeError("%s\n", err)
	}

	screen.Header = &term.TextLine{
		Content:    "    Pulls: Pull Request Management",
		Forground:  termbox.ColorWhite,
		Background: termbox.ColorYellow,
	}
	cursor := term.NewCursor(screen)

	pulls, err := m.GetPullRequests("open")
	if err != nil {
		writeError("%s\n", err)
	}

	for _, pr := range pulls {
		cells := &term.CellLine{make([]*term.Cell, 3)}

		cells.Cells[0] = term.NewCell(strconv.Itoa(pr.Number), termbox.ColorDefault, termbox.ColorBlack)
		cells.Cells[1] = term.NewCell(pr.Title, termbox.ColorDefault, termbox.ColorBlack)
		cells.Cells[1].Width = 50
		cells.Cells[2] = term.NewCell(pr.CreatedAt.Format(defaultTimeFormat), termbox.ColorDefault, termbox.ColorBlack)

		screen.Lines = append(screen.Lines, cells)
	}

	if err := screen.Display(); err != nil {
		writeError("%s\n", err)
	}

	// Main event loop
	for {
		switch ev := term.Event(); ev.Type {
		case termbox.EventError:
			err = ev.Err
			goto exit
		case termbox.EventResize:
			err = screen.Resize()
			if err != nil {
				goto exit
			}
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc, termbox.KeyCtrlQ:
				goto exit
			case termbox.KeyArrowDown:
				cursor.Down()
			case termbox.KeyArrowUp:
				cursor.Up()
			case termbox.KeyEnter:
				cursor.Select()
			}
		}
	}
exit:
	screen.Close()
	if err != nil {
		writeError("%s\n", err)
	}
}

func listOpenPullsCmd(c *cli.Context) {
	prs, err := m.GetPullRequests("open")
	filters := &pulls.ShowFilters{
		NoMerge:  c.Bool("no-merge"),
		FromUser: c.String("user"),
	}
	if filters.NoMerge || filters.FromUser != "" {
		prs, err = m.FilterPullRequests(prs, filters)
		if err != nil {
			writeError("Error getting pull requests %s", err)
		}
	}
	fmt.Printf("%c[2K\r", 27)
	displayPullRequests(prs)
}

func listClosedPullsCmd(c *cli.Context) {
	pulls, err := m.GetPullRequests("closed")
	if err != nil {
		writeError("Error getting pull requests %s", err)
	}
	displayPullRequests(pulls)
}

func displayPullRequests(pulls []*gh.PullRequest) {
	w := newTabwriter()
	for _, p := range pulls {
		fmt.Fprintf(w, "%d\t%s\t%s\n", p.Number, truncate(p.Title), p.CreatedAt.Format(defaultTimeFormat))
	}

	if err := w.Flush(); err != nil {
		writeError("%s", err)
	}
}

func showPullRequestCmd(c *cli.Context) {
	pr, err := m.GetPullRequest(c.Args()[0])
	if err != nil {
		writeError("%s", err)
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
		writeError("%s", err)
	}
	fmt.Fprintf(os.Stdout, "Name: %s\nForks: %d\nStars: %d\nIssues: %d\n", r.Name, r.Forks, r.Watchers, r.OpenIssues)
}

func authCmd(c *cli.Context) {
	if token := c.String("add"); token != "" {
		if err := saveConfig(Config{token}); err != nil {
			writeError("%s", err)
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
	number := c.Args()[0]
	if c.Bool("add") {
		comment := c.Args()[1]
		cmt, err := m.AddComment(number, comment)
		if err != nil {
			writeError("%s\n", err)
		}
		fmt.Fprintf(os.Stdout, "Comment added at %s\n", cmt.CreatedAt.Format(defaultTimeFormat))
		return
	} else {
		comments, err := m.GetComments(number)
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

func loadCommands(app *cli.App) {
	app.Commands = []cli.Command{
		{
			Name:      "open",
			ShortName: "o",
			Usage:     "List all open pull requests for the current repository",
			Action:    listOpenPullsCmd,
			Flags: []cli.Flag{
				cli.BoolFlag{"no-merge", "display only prs that cannot be merged"},
				cli.StringFlag{"user", "", "display only prs from <user>"},
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
			Usage:     "Show and manage comments for a pull request",
			Action:    manageCommentsCmd,
			Flags: []cli.Flag{
				cli.BoolFlag{"add", "add a comment to the pull request"},
			},
		},
		{
			Name:   "merge",
			Usage:  "Merge a pull request",
			Action: mergeCmd,
			Flags: []cli.Flag{
				cli.StringFlag{"m", "", "commit message for merge"},
			},
		},
		{
			Name:   "interactive",
			Action: displayInteractiveCmd,
			Usage:  "Display an interactive screen",
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
