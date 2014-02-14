package gordon

type ContributorStats struct {
	Name      string
	Additions int
	Deletions int
	Commits   int
}

type ByAdditions []ContributorStats
type ByDeletions []ContributorStats
type ByCommits []ContributorStats

func (a ByAdditions) Len() int           { return len(a) }
func (a ByAdditions) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByAdditions) Less(i, j int) bool { return a[j].Additions < a[i].Additions }

func (a ByDeletions) Len() int           { return len(a) }
func (a ByDeletions) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDeletions) Less(i, j int) bool { return a[j].Deletions < a[i].Deletions }

func (a ByCommits) Len() int           { return len(a) }
func (a ByCommits) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCommits) Less(i, j int) bool { return a[j].Commits < a[i].Commits }
