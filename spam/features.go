package spam

import "time"

var AssocToClass = map[string]int{
	"NONE":                   0,
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

	// AccountAge is the number of days account was open
	// when the issue was posted
	AccountAge int

	// IsBioSet is 1 if author has a bio, else 0
	IsBioSet int

	// Number of chars in the Issue content
	TitleLen int
	BodyLen  int

	// The max similarity score between the issue and the repo's issue templates
	TemplateScore int

	Followers int
	Following int

	// IsSpam is 1 if issue was spam, else 0
	IsSpam int
}

func GetFeatures(issue Issue, author User, templates []string) Features {
	issueCreated, _ := time.Parse(time.RFC3339, issue.CreatedAt)
	acctCreated, _ := time.Parse(time.RFC3339, author.CreatedAt)
	acctAge := int(issueCreated.Sub(acctCreated).Hours() / 24)

	simScore := MaxSimScore(issue.Body, templates)

	feats := Features{
		Association:   AssocToClass[issue.AuthorAssociation],
		Following:     author.Following,
		Followers:     author.Followers,
		AccountAge:    acctAge,
		Contributions: author.TotalContributions,
		AuthorRepos:   author.ReposContributed,
		TitleLen:      len(issue.Title),
		BodyLen:       len(issue.Body),
		TemplateScore: simScore,
	}

	if author.Bio != "" {
		feats.IsBioSet = 1
	}
	if issue.IsSpam {
		feats.IsSpam = 1
	}
	return feats
}
