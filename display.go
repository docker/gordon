package pulls

import (
	"fmt"
	"github.com/aybabtme/color/brush"
	"github.com/codegangsta/cli"
	gh "github.com/crosbymichael/octokat"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

const (
	defaultTimeFormat = time.RFC822
)

func newTabwriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 8, 1, 3, ' ', 0)
}

func truncate(s string) string {
	if len(s) > 30 {
		s = s[:30] + "..."
	}
	return s
}

func DisplayPullRequests(c *cli.Context, pulls []*gh.PullRequest, notrunc bool) {
	w := newTabwriter()
	fmt.Fprintf(w, "NUMBER\tTITLE\tCREATED AT")
	if c.Bool("lgtm") {
		fmt.Fprintf(w, "\tLGTM")
	}
	fmt.Fprintf(w, "\n")
	for _, p := range pulls {
		if !notrunc {
			p.Title = truncate(p.Title)
		}
		fmt.Fprintf(w, "%d\t%s\t%s", p.Number, p.Title, p.CreatedAt.Format(defaultTimeFormat))
		if c.Bool("lgtm") {
			fmt.Fprintf(w, "\t%d", p.ReviewComments)
		}
		fmt.Fprintf(w, "\n")
	}

	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

func DisplayPullRequest(pr *gh.PullRequest) {
	fmt.Fprint(os.Stdout, brush.Green("Pull Request:"), "\n")
	fmt.Fprintf(os.Stdout, "No: %d\nTitle: %s\n\n", pr.Number, pr.Title)

	lines := strings.Split(pr.Body, "\n")
	for i, l := range lines {
		lines[i] = "\t" + l
	}
	fmt.Fprintf(os.Stdout, "Description:\n\n%s\n\n", strings.Join(lines, "\n"))
}

func DisplayComments(comments []gh.Comment) {
	fmt.Fprintln(os.Stdout, "Comments:\n")
	for _, c := range comments {
		fmt.Fprintf(os.Stdout, "@%s %s\n%s\n", brush.Red(c.User.Login), c.CreatedAt.Format(defaultTimeFormat), c.Body)
		fmt.Fprint(os.Stdout, "\n\n")
	}
}

func DisplayCommentAdded(cmt gh.Comment) {
	fmt.Fprintf(os.Stdout, "Comment added at %s\n", cmt.CreatedAt.Format(defaultTimeFormat))
}
