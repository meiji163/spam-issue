package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cli/go-gh"
	"github.com/meiji163/gh-spam/classify"
	"github.com/meiji163/gh-spam/spam"
	"github.com/sjwhitworth/golearn/base"
	"github.com/sjwhitworth/golearn/ensemble"
	"github.com/sjwhitworth/golearn/evaluation"
	"github.com/spf13/cobra"
)

const numTrees = 61

func main() {
	cmd := rootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

type SpamOpts struct {
	Numbers []int
	RepoArg string
	Repo    string
	Owner   string
	DataDir string
}

func rootCmd() *cobra.Command {
	opts := &SpamOpts{Numbers: []int{}}
	cmd := &cobra.Command{
		Use:   "spam",
		Short: "Classify spam issues",
		Args:  cobra.ExactArgs(1),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.RepoArg == "" {
				repo, err := gh.CurrentRepository()
				if err != nil {
					return fmt.Errorf("No repository argument")
				}
				opts.Repo = repo.Name()
				opts.Owner = repo.Owner()
			} else {
				ownerRepo := strings.Split(opts.RepoArg, "/")
				if len(ownerRepo) != 2 {
					return fmt.Errorf("Invalid repository argument")
				}
				opts.Owner = ownerRepo[0]
				opts.Repo = ownerRepo[1]
			}
			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.RepoArg, "repo", "R", "", "specify the repository in OWNER/REPO format")

	downloadCmd := &cobra.Command{
		Use:   "download",
		Short: "Download issue dataset from a GitHub repository",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDownload(opts)
		},
	}

	classifyCmd := &cobra.Command{
		Use:   "classify <number>",
		Short: "Classify issues as spam. Accepts one or more issue numbers",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				num, err := strconv.Atoi(arg)
				if err != nil {
					return fmt.Errorf("Invalid issue number %s", arg)
				}
				opts.Numbers = append(opts.Numbers, num)
			}

			return runClassify(opts)
		},
	}

	cmd.AddCommand(downloadCmd, classifyCmd)
	return cmd
}

func runClassify(opts *SpamOpts) error {
	dataPath := fmt.Sprintf("%s-%s.csv", opts.Owner, opts.Repo)
	if _, err := os.Stat(dataPath); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("data for %s/%s not found", opts.Owner, opts.Repo)
	}

	tree := ensemble.NewRandomForest(numTrees, 9)

	dataset, err := base.ParseCSVToInstances(dataPath, true)
	if err != nil {
		return err
	}

	if err = tree.Fit(dataset); err != nil {
		return err
	}

	templates, err := spam.GetTemplates(opts.Owner, opts.Repo)
	if err != nil {
		return err
	}

	issues := []spam.Issue{}
	feats := []spam.Features{}
	for _, num := range opts.Numbers {
		issue, err := spam.GetIssueByNumber(opts.Owner, opts.Repo, num)
		if err != nil {
			return err
		}
		issues = append(issues, issue)

		username := issue.Author.Login
		author, err := spam.GetUserStats(username)
		if err != nil {
			return fmt.Errorf("Error getting user stats for %s: %s", username, err)
		}

		feat := spam.ExtractFeatures(issue, author, templates)
		feats = append(feats, feat)
	}

	instance := classify.FeaturesToInstances(feats)

	pred, err := tree.Predict(instance)
	if err != nil {
		return err
	}

	for i := 0; i < len(issues); i++ {
		label := "not spam"
		if pred.RowString(i) != "0" {
			label = "spam"
		}

		fmt.Printf("#%d: %s\n", opts.Numbers[i], label)
	}
	return nil
}

func runDownload(opts *SpamOpts) error {
	dataPath := fmt.Sprintf("%s-%s.csv", opts.Owner, opts.Repo)
	if _, err := os.Stat(dataPath); errors.Is(err, os.ErrNotExist) {
		_, err = os.Create(dataPath)
		if err != nil {
			return err
		}
	}

	feats, err := spam.MakeDataset(spam.MakeOpts{Owner: opts.Owner, Repo: opts.Repo})
	if err != nil {
		return err
	}

	dataset := classify.FeaturesToInstances(feats)
	if err := base.SerializeInstancesToCSV(dataset, dataPath); err != nil {
		return err
	}

	tree := ensemble.NewRandomForest(numTrees, 9)
	if err = tree.Fit(dataset); err != nil {
		return err
	}

	predictions, err := tree.Predict(dataset)
	if err != nil {
		return err
	}

	cm, err := evaluation.GetConfusionMatrix(dataset, predictions)
	if err != nil {
		return err
	}
	fmt.Println(evaluation.GetSummary(cm))

	// serialize model with gob
	modelPath := fmt.Sprintf("%s-%s.gob", opts.Owner, opts.Repo)
	err = classify.WriteGob(modelPath, *tree)
	return err
}
