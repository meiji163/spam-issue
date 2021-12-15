package main

import (
	"errors"
	"fmt"
	"log"
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

func main() {
	cmd := rootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

type SpamOpts struct {
	Number  int
	RepoArg string
	Repo    string
	Owner   string
}

func rootCmd() *cobra.Command {
	opts := &SpamOpts{}
	cmd := &cobra.Command{
		Use:  "spam",
		Args: cobra.ExactArgs(1),
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

	cmd.PersistentFlags().StringVarP(&opts.RepoArg, "repo", "R", "", "specify the repository")

	trainCmd := &cobra.Command{
		Use:  "train",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTrain(opts)
		},
	}

	classifyCmd := &cobra.Command{
		Use:  "classify",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			num, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("Invalid issue number %s", args[0])
			}
			opts.Number = num

			return runClassify(opts)
		},
	}

	cmd.AddCommand(trainCmd, classifyCmd)
	return cmd
}

func runClassify(opts *SpamOpts) error {
	issue, err := spam.GetIssueByNumber(opts.Owner, opts.Repo, opts.Number)
	if err != nil {
		return err
	}

	fmt.Println(issue)
	return nil
}

func runTrain(opts *SpamOpts) error {
	dataPath := fmt.Sprintf("%s-%s.csv", opts.Owner, opts.Repo)

	if _, err := os.Stat(dataPath); errors.Is(err, os.ErrNotExist) {
		_, err = os.Create(dataPath)
		if err != nil {
			return err
		}
	}

	makeOpts := spam.MakeOpts{Owner: opts.Owner, Repo: opts.Repo}
	feats, err := spam.MakeDataset(makeOpts)
	if err != nil {
		return err
	}
	log.Println("Done making dataset")

	dataset := classify.FeaturesToInstances(feats)

	if err := base.SerializeInstancesToCSV(dataset, dataPath); err != nil {
		return err
	}

	tree := ensemble.NewRandomForest(71, 9)
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

	modelPath := fmt.Sprintf("%s-%s.gob", opts.Owner, opts.Repo)
	err = classify.WriteGob(modelPath, tree)
	return err
}
