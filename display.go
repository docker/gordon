package gordon

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/codegangsta/cli"
	gh "github.com/crosbymichael/octokat"
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
	fmt.Fprintf(w, "NUMBER\tSHA\tLAST UPDATED\tCONTRIBUTOR\tASSIGNEE\tTITLE")
	if c.Bool("lgtm") {
		fmt.Fprintf(w, "\tLGTM")
	}
	fmt.Fprintf(w, "\n")
	for _, p := range pulls {
		if !notrunc {
			p.Title = truncate(p.Title)
		}
		var assignee string
		if p.Assignee != nil {
			assignee = p.Assignee.Login
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s", p.Number, p.Head.Sha[:8], HumanDuration(time.Since(p.UpdatedAt)), p.User.Login, assignee, p.Title)
		if c.Bool("lgtm") {
			lgtm := strconv.Itoa(p.ReviewComments)
			if p.ReviewComments >= 2 {
				lgtm = Green(lgtm)
			} else if p.ReviewComments == 0 {
				lgtm = DarkRed(lgtm)
			} else {
				lgtm = DarkYellow(lgtm)
			}
			fmt.Fprintf(w, "\t%s", lgtm)
		}
		fmt.Fprintf(w, "\n")
	}

	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

func DisplayReviewers(c *cli.Context, reviewers map[string][]string) {
	w := newTabwriter()
	fmt.Fprintf(w, "FILE\tREVIEWERS")
	fmt.Fprintf(w, "\n")
	for file, fileReviewers := range reviewers {
		var usernames bytes.Buffer
		for _, reviewer := range fileReviewers {
			usernames.WriteString(reviewer)
			usernames.WriteString(", ")
		}
		usernames.Truncate(usernames.Len() - 2)
		fmt.Fprintf(w, "%s\t%s\n", file, usernames.String())
	}
	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

func DisplayContributors(c *cli.Context, contributors []*gh.Contributor) {
	var (
		w                 = newTabwriter()
		contributorsStats []ContributorStats
	)

	for _, contrib := range contributors {
		contribStats := ContributorStats{}
		contribStats.Name = contrib.Author.Login
		for _, week := range contrib.Weeks {
			contribStats.Additions += week.Additions
			contribStats.Deletions += week.Deletions
			contribStats.Commits += week.Commits
		}
		contributorsStats = append(contributorsStats, []ContributorStats{contribStats}...)
	}
	if c.Bool("additions") {
		sort.Sort(ByAdditions(contributorsStats))
	} else if c.Bool("deletions") {
		sort.Sort(ByDeletions(contributorsStats))
	} else if c.Bool("commits") {
		sort.Sort(ByCommits(contributorsStats))
	} else {
		// Sort by default by Commits
		sort.Sort(ByCommits(contributorsStats))
	}
	topN := c.Int("top")
	fmt.Fprintf(w, "CONTRIBUTOR\tADDITIONS\tDELETIONS\tCOMMITS")
	fmt.Fprintf(w, "\n")
	for i := 0; i < len(contributorsStats) && i < topN; i++ {
		fmt.Fprintf(w, "%s\t%d\t%d\t%d", contributorsStats[i].Name,
			contributorsStats[i].Additions,
			contributorsStats[i].Deletions,
			contributorsStats[i].Commits)
		fmt.Fprintf(w, "\n")
	}

	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

func DisplayPullRequest(pr *gh.PullRequest) {
	fmt.Fprint(os.Stdout, fmt.Sprintf("Pull Request from: %s", Green("@"+pr.User.Login)), "\n")
	fmt.Printf("No: %d\nSha: %s\nTitle: %s\n", pr.Number, pr.Head.Sha, pr.Title)

	if pr.Merged {
		fmt.Fprintf(os.Stdout, "\nMerged by: %s\nMerged at: %s\nMerge Commit: %s\n\n", Yellow("@"+pr.MergedBy.Login), Yellow(pr.MergedAt.Format(time.RubyDate)), Yellow(pr.MergeCommitSha))
	} else {
		m := fmt.Sprintf("%t", pr.Mergeable)
		if pr.Mergeable {
			fmt.Fprintf(os.Stdout, "Mergeable: %s", Green(m))
		} else {
			fmt.Fprintf(os.Stdout, "Mergeable: %s", Red(m))
		}
	}
	fmt.Fprint(os.Stdout, "\n")

	lines := strings.Split(pr.Body, "\n")
	for i, l := range lines {
		lines[i] = "\t" + l
	}
	fmt.Printf("Description:\n\n%s\n\n", strings.Join(lines, "\n"))
	fmt.Printf("\n\n")

	DisplayComments(pr.CommentsBody)
}

func DisplayComments(comments []gh.Comment) {
	fmt.Fprintln(os.Stdout, "Comments:")
	for _, c := range comments {
		fmt.Printf("<%s\n@%s %s\n%s\n%s>", strings.Repeat("=", 79), Red(c.User.Login), c.CreatedAt.Format(defaultTimeFormat), strings.Replace(c.Body, "LGTM", fmt.Sprintf("%s", Green("LGTM")), -1), strings.Repeat("=", 79))
		fmt.Fprint(os.Stdout, "\n\n")
	}
}

func DisplayCommentAdded(cmt gh.Comment) {
	fmt.Printf("Comment added at %s\n", cmt.CreatedAt.Format(defaultTimeFormat))
}

func printIssue(c *cli.Context, w *tabwriter.Writer, number int, updatedAt time.Time, login string, title string, comments int) {
	fmt.Fprintf(w, "%d\t%s\t%s\t%s", number, HumanDuration(time.Since(updatedAt)), login, title)
	if c.Int("votes") > 0 {
		votes := strconv.Itoa(comments)
		if comments >= 2 {
			votes = Green(votes)
		}
		fmt.Fprintf(w, "\t%s", votes)
	}
	fmt.Fprintf(w, "\n")
}

// Display Issues prints `issues` to standard output in a human-friendly tabulated format.
func DisplayIssues(c *cli.Context, v interface{}, notrunc bool) {
	w := newTabwriter()
	fmt.Fprintf(w, "NUMBER\tLAST UPDATED\tASSIGNEE\tTITLE")
	if c.Int("votes") > 0 {
		fmt.Fprintf(w, "\tVOTES")
	}
	fmt.Fprintf(w, "\n")

	switch issues := v.(type) {
	case []*gh.Issue:
		for _, p := range issues {
			printIssue(c, w, p.Number, p.UpdatedAt, p.Assignee.Login, p.Title, p.Comments)
		}
	case []*gh.SearchItem:
		for _, p := range issues {
			printIssue(c, w, p.Number, p.UpdatedAt, p.Assignee.Login, p.Title, p.Comments)
		}
	}
	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

func DisplayIssue(issue *gh.Issue, comments []gh.Comment) {
	fmt.Fprint(os.Stdout, Green("Issue:"), "\n")
	fmt.Printf("No: %d\nTitle: %s\n\n", issue.Number, issue.Title)

	lines := strings.Split(issue.Body, "\n")
	for i, l := range lines {
		lines[i] = "\t" + l
	}
	fmt.Printf("Description:\n\n%s\n\n", strings.Join(lines, "\n"))
	fmt.Printf("\n\n")

	DisplayComments(comments)
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

func DisplayPatch(r io.Reader) error {
	s := bufio.NewScanner(r)
	for s.Scan() {
		if err := s.Err(); err != nil {
			return err
		}
		t := s.Text()

		switch t[0] {
		case '-':
			fmt.Fprintln(os.Stdout, Red(t))
		case '+':
			fmt.Fprintln(os.Stdout, Green(t))
		default:
			fmt.Fprintln(os.Stdout, t)
		}
	}
	return nil
}
