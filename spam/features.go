package spam

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/cheggaaa/pb/v3"
)

var AssocToClass = map[string]int{
	"NONE":                   0,
	"FIRST_TIMER":            0,
	"FIRST_TIME_CONTRIBUTOR": 1,
	"COLLABORATOR":           2,
	"CONTRIBUTOR":            3,
	"MEMBER":                 4,
	"OWNER":                  4,
}

type MakeOpts struct {
	Owner   string
	Repo    string
	Limit   int
	Verbose bool
}

func MakeDataset(opts MakeOpts) ([]Features, error) {
	log.Printf("Downloading issues for %s/%s\n", opts.Owner, opts.Repo)
	issues, err := downloadIssues(opts.Owner, opts.Repo)
	if err != nil {
		return nil, err
	}

	authors := map[string]User{}
	feats := []Features{}

	// fetch issue templates for matching
	templates, err := GetTemplates(opts.Owner, opts.Repo)
	if err != nil {
		return nil, err
	}

	log.Printf("%d templates\n", len(templates))

	bar := pb.StartNew(len(issues))

	// get the issue author's stats to compute dataset features
	log.Println("Processing Issues")
	for _, issue := range issues {
		username := issue.Author.Login
		author, ok := authors[username]
		if !ok {
			author, err = GetUserStats(username)
			if err != nil {
				continue
			}
			authors[username] = author
		}
		feat := ExtractFeatures(issue, author, templates)
		feats = append(feats, feat)
		bar.Increment()
	}
	bar.Finish()

	return feats, nil
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

// ExtractFeatures gets numeric features from issue for classification
func ExtractFeatures(issue Issue, author User, templates []string) Features {
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

	// assume contributors never post spam
	if issue.IsSpam && feats.Association < 3 {
		feats.IsSpam = 1
	}
	return feats
}

func downloadIssues(owner, repo string) ([]Issue, error) {
	issues, err := GetNonSpam(owner, repo)
	if err != nil {
		return nil, err
	}

	if len(issues) == 0 {
		return nil, fmt.Errorf("No issues found in %s/%s", owner, repo)
	}

	spamIssues, err := GetSpam(owner, repo)
	if err != nil {
		return nil, err
	}

	if len(spamIssues) == 0 {
		return nil, fmt.Errorf("No spam issues found in %s/%s", owner, repo)
	}

	for _, spamIssue := range spamIssues {
		spamIssue.IsSpam = true
		issues = append(issues, spamIssue)
	}
	sort.Sort(byNumber(issues))
	return issues, nil
}

type byNumber []Issue

func (l byNumber) Len() int {
	return len(l)
}

func (l byNumber) Less(i, j int) bool {
	return l[i].Number < l[j].Number
}

func (l byNumber) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
