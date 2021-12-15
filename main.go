package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
		os.Exit(1)
	}
}

type SpamOpts struct {
	Numbers   []int
	RepoArg   string
	Repo      string
	Owner     string
	DataPath  string
	ModelPath string
	Limit     int
	Verbose   bool
}

func rootCmd() *cobra.Command {
	opts := &SpamOpts{Numbers: []int{}}
	cmd := &cobra.Command{
		Use:   "spam",
		Short: "Classify GitHub issues as spam.",
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

			opts.DataPath = filepath.Join("data", fmt.Sprintf("%s-%s.csv", opts.Owner, opts.Repo))
			opts.ModelPath = filepath.Join("data", fmt.Sprintf("%s-%s.gob", opts.Owner, opts.Repo))
			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.RepoArg, "repo", "R", "", "specify the repository in OWNER/REPO format")
	cmd.PersistentFlags().BoolVarP(&opts.Verbose, "verbose", "v", false, "verbose mode")

	downloadCmd := &cobra.Command{
		Use:   "download",
		Short: "Download issue dataset from a GitHub repository",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDownload(opts)
		},
	}
	downloadCmd.Flags().IntVarP(&opts.Limit, "limit", "L", 600, "max number of issues to download")

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
	if _, err := os.Stat(opts.ModelPath); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("model for %s/%s not found", opts.Owner, opts.Repo)
	}

	tree := ensemble.NewRandomForest(numTrees, 9)
	if err := tree.Load(opts.ModelPath); err != nil {
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

	instances := classify.FeaturesToInstances(feats)
	pred, err := tree.Predict(instances)
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
	datasetFound := true
	_, err := os.Stat(opts.DataPath)
	if errors.Is(err, os.ErrNotExist) {
		_ = os.Mkdir("data", 0700)
		_, err = os.Create(opts.DataPath)
		if err != nil {
			return err
		}
		datasetFound = false
	}

	var dataset *base.DenseInstances
	if !datasetFound {
		makeOpts := spam.MakeOpts{
			Owner:   opts.Owner,
			Repo:    opts.Repo,
			Limit:   opts.Limit,
			Verbose: opts.Verbose}
		feats, err := spam.MakeDataset(makeOpts)
		if err != nil {
			return err
		}
		dataset = classify.FeaturesToInstances(feats)

		if err := base.SerializeInstancesToCSV(dataset, opts.DataPath); err != nil {
			return err
		}

	} else {
		dataset, err = base.ParseCSVToInstances(opts.DataPath, true)
		if err != nil {
			return err
		}
	}

	tree := ensemble.NewRandomForest(numTrees, 9)
	if err := tree.Fit(dataset); err != nil {
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

	// serialize model
	return tree.Save(opts.ModelPath)
}
