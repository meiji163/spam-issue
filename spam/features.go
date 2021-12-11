package spam

var AssocToClass = map[string]int{
	"FIRST_TIMER":            0,
	"FIRST_TIME_CONTRIBUTOR": 1,
	"COLLABORATOR":           2,
	"CONTRIBUTOR":            3,
	"MEMBER":                 4,
	"OWNER":                  4,
}

type Features struct {
	// A class label for author's association to the repo
	Association int

	// Contributions is the number of author's contributions on GitHub
	// in the last year, as shown on their profile
	Contributions int

	// AuthorRepos is the number of repos author contributed to
	AuthorRepos int

	// AccountAge is the days since account was created
	AccountAge int

	// IsBioSet is 1 if author has a bio, else 0
	IsBioSet int

	Followers int
	Following int

	// IsSpam is 1 if issue was spam, else 0
	IsSpam int
}

func GetFeatures(issue *Issue, author *User) Features {
	feats := Features{
		Association:   AssocToClass[issue.AuthorAssociation],
		Following:     author.Following,
		Followers:     author.Followers,
		AccountAge:    int(author.Age.Hours()) / 24,
		Contributions: author.TotalContributions,
		AuthorRepos:   author.ReposContributed,
	}

	if author.Bio != "" {
		feats.IsBioSet = 1
	}
	if issue.IsSpam {
		feats.IsSpam = 1
	}
	return feats
}
