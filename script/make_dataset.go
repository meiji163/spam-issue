// Downloads issues from a repo for training a spam classifier

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/meiji163/gh-spam/spam"
)

type byNumber []spam.Issue

func (l byNumber) Len() int {
	return len(l)
}

func (l byNumber) Less(i, j int) bool {
	return l[i].Number < l[j].Number
}

func (l byNumber) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func downloadIssues(owner, repo string) ([]spam.Issue, error) {
	issues, err := spam.GetNonSpam(owner, repo)
	if err != nil {
		return nil, err
	}

	spamIssues, err := spam.GetSpam(owner, repo)
	log.Printf("(%d) spam issues\n", len(spamIssues))
	if err != nil {
		return nil, err
	}

	for _, spamIssue := range spamIssues {
		spamIssue.IsSpam = true
		issues = append(issues, spamIssue)
	}
	sort.Sort(byNumber(issues))
	return issues, nil
}

func writeToCSV(fname string, feats []spam.Features) error {
	f, err := os.Create(fname)
	if err != nil {
		return err
	}

	defer f.Close()

	header := "association,contributions,repos,age,followers,following,body_len,title_len,sim_score,is_spam\n"
	f.WriteString(header)

	for _, feat := range feats {
		f.WriteString(
			fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d,%d,%d,%d\n",
				feat.Association,
				feat.Contributions,
				feat.AuthorRepos,
				feat.AccountAge,
				feat.Followers,
				feat.Following,
				feat.BodyLen,
				feat.TitleLen,
				feat.TemplateScore,
				feat.IsSpam))
	}
	return nil
}

func help() {
	fmt.Printf("Usage: %s <owner/repo>\n", filepath.Base(os.Args[0]))
	os.Exit(1)
}

func main() {
	if len(os.Args) != 2 {
		help()
	}
	fullName := os.Args[1]
	names := strings.Split(fullName, "/")
	if len(names) != 2 {
		help()
	}
	owner := names[0]
	repo := names[1]

	log.Printf("Downloading issues for %s/%s\n", owner, repo)
	issues, err := downloadIssues(owner, repo)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Downloaded (%d) issues\n", len(issues))

	log.Println("Creating dataset")
	authors := map[string]spam.User{}
	feats := []spam.Features{}

	// fetch issue templates for matching
	templates, err := spam.GetTemplates(owner, repo)
	log.Printf("%d templates\n", len(templates))
	if err != nil {
		log.Fatal(err)
	}

	// get the issue author's stats to compute data features
	for i, issue := range issues {
		username := issue.Author.Login

		author, ok := authors[username]
		if !ok {
			author, err = spam.GetUserStats(username)
			if err != nil {
				log.Fatal(err)
			}
			authors[username] = author
		}
		feat := spam.GetFeatures(issue, author, templates)
		feats = append(feats, feat)

		if i%20 == 19 {
			log.Printf("%d/%d processed\n", i+1, len(issues))
		}
	}

	// save dataset to csv
	filename := fmt.Sprintf("%s-%s_data.csv", owner, repo)
	if err := writeToCSV(filename, feats); err != nil {
		log.Fatal(err)
	}
}
