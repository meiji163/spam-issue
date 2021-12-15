# gh-spam
gh spam is a classifier for spam issues on GitHub that achieves high accuracy with only public user stats and issue metadata.    
It uses [go-gh](https://github.com/cli/go-gh) for GitHub API and [golearn](https://github.com/sjwhitworth/golearn) for classification.

# usage
First download a dataset of issues from your repo, then you can train a classifier for inference.   
Here is an example with the [cli/cli](https://github.com/cli/cli) repo:

```shell
$ gh-spam download -R cli/cli

$ gh-spam classify -R cli/cli 4894
#4894: spam
```

You can also pass multiple issue numbers for classification.
```shell
$ gh-spam classify -R cli/cli 4913 4907 4906 4894
#4913: not spam
#4907: not spam
#4906: not spam
#4894: spam
```

# details
By default, the classifier is a random forest. Any classifier from golearn can easily be substituted.    
The main inputs are:
- author's association with the repo
- age of author's account
- number of the author's contributions on GitHub
- the length of the issue title and body
- a matching score between the issue and the repo's issue templates


Here is the classifier accuracy on the cli/cli data
```
Reference Class	True Positives	False Positives	True Negatives	Precision	Recall	F1 Score
---------------	--------------	---------------	--------------	---------	------	--------
0		544		3		219		0.9945		0.9963	0.9954
1		219		2		544		0.9910		0.9865	0.9887
Overall accuracy: 0.9935
```
