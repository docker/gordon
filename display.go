package pulls

import (
	"fmt"
	"github.com/aybabtme/color/brush"
	"github.com/codegangsta/cli"
	gh "github.com/crosbymichael/octokat"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

const (
	defaultTimeFormat = time.RFC822
	truncSize         = 80
)

func newTabwriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 8, 1, 3, ' ', 0)
}

func truncate(s string) string {
	if len(s) > truncSize {
		s = s[:truncSize] + "..."
	}
	return s
}

func DisplayPullRequests(c *cli.Context, pulls []*gh.PullRequest, notrunc bool) {
	w := newTabwriter()
	fmt.Fprintf(w, "NUMBER\tLAST UPDATED\tTITLE")
	if c.Bool("lgtm") {
		fmt.Fprintf(w, "\tLGTM")
	}
	fmt.Fprintf(w, "\n")
	for _, p := range pulls {
		if !notrunc {
			p.Title = truncate(p.Title)
		}
		fmt.Fprintf(w, "%d\t%s\t%s", p.Number, HumanDuration(time.Since(p.UpdatedAt)), p.Title)
		if c.Bool("lgtm") {
			lgtm := strconv.Itoa(p.ReviewComments)
			if p.ReviewComments >= 2 {
				lgtm = brush.Green(lgtm).String()
			}
			fmt.Fprintf(w, "\t%s", lgtm)
		}
		fmt.Fprintf(w, "\n")
	}

	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

func DisplayPullRequest(pr *gh.PullRequest, comments []gh.Comment) {
	fmt.Fprint(os.Stdout, brush.Green("Pull Request:"), "\n")
	fmt.Fprintf(os.Stdout, "No: %d\nTitle: %s\n\n", pr.Number, pr.Title)

	lines := strings.Split(pr.Body, "\n")
	for i, l := range lines {
		lines[i] = "\t" + l
	}
	fmt.Fprintf(os.Stdout, "Description:\n\n%s\n\n", strings.Join(lines, "\n"))
	fmt.Fprintf(os.Stdout, "\n\n")

	DisplayComments(comments)
}

func DisplayComments(comments []gh.Comment) {
	fmt.Fprintln(os.Stdout, "Comments:")
	for _, c := range comments {
		fmt.Fprintf(os.Stdout, "<%s\n@%s %s\n%s\n%s>", strings.Repeat("=", 79), brush.Red(c.User.Login), c.CreatedAt.Format(defaultTimeFormat), strings.Replace(c.Body, "LGTM", fmt.Sprintf("%s", brush.Green("LGTM")), -1), strings.Repeat("=", 79))
		fmt.Fprint(os.Stdout, "\n\n")
	}
}

func DisplayCommentAdded(cmt gh.Comment) {
	fmt.Fprintf(os.Stdout, "Comment added at %s\n", cmt.CreatedAt.Format(defaultTimeFormat))
}

// DisplayIssues prints `issues` to standard output in a human-friendly tabulated format.
func DisplayIssues(c *cli.Context, issues []*gh.Issue, notrunc bool) {
	w := newTabwriter()
	fmt.Fprintf(w, "NUMBER\tLAST UPDATED\tASSIGNEE\tTITLE")
	if c.Int("votes") > 0 {
		fmt.Fprintf(w, "\tVOTES")
	}
	fmt.Fprintf(w, "\n")
	for _, p := range issues {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s", p.Number, HumanDuration(time.Since(p.UpdatedAt)), p.Assignee.Login, p.Title)
		if c.Int("votes") > 0 {
			votes := strconv.Itoa(p.Comments)
			if p.Comments >= 2 {
				votes = brush.Green(votes).String()
			}
			fmt.Fprintf(w, "\t%s", votes)
		}
		fmt.Fprintf(w, "\n")
	}

	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

// HumanDuration returns a human-readable approximation of a duration
// This function is taken from the Docker project, and slightly modified
// to cap units at days.
// (eg. "About a minute", "4 hours ago", etc.)
// (c) 2013 Docker, inc. and the Docker authors (http://docker.io)
func HumanDuration(d time.Duration) string {
	if seconds := int(d.Seconds()); seconds < 1 {
		return "Less than a second"
	} else if seconds < 60 {
		return fmt.Sprintf("%d seconds", seconds)
	} else if minutes := int(d.Minutes()); minutes == 1 {
		return "About a minute"
	} else if minutes < 60 {
		return fmt.Sprintf("%d minutes", minutes)
	} else if hours := int(d.Hours()); hours == 1 {
		return "About an hour"
	} else if hours < 48 {
		return fmt.Sprintf("%d hours", hours)
	}
	return fmt.Sprintf("%d days", int(d.Hours()/24))
}
