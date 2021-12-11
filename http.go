package spam

import (
	"fmt"
	"time"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
)

type User struct {
	Name               string
	Age                time.Duration
	Bio                string
	Followers          int
	Following          int
	TotalContributions int
	ReposContributed   int
}

type Issue struct {
	Number            int
	Title             string
	Body              string
	Author            struct{ Login string }
	AuthorAssociation string
}

func getUserStats(username string) (*User, error) {
	opts := &api.ClientOptions{EnableCache: true}
	client, err := gh.GQLClient(opts)
	if err != nil {
		return nil, err
	}

	query := `query GetUserStats($username: String!) {
  user(login: $username) {
    createdAt
    bio
    followers{ totalCount }
    following{ totalCount }
    contributionsCollection {
      contributionCalendar { totalContributions }
    }
    repositoriesContributedTo(
		first:100, 
		contributionTypes: [COMMIT, ISSUE, PULL_REQUEST], 
		orderBy: {field: UPDATED_AT,direction: DESC}){
      totalCount
    }
  }
}`

	variables := map[string]interface{}{"username": username}
	resp := struct {
		User struct {
			CreatedAt               string
			Bio                     string
			Followers               struct{ TotalCount int }
			Following               struct{ TotalCount int }
			ContributionsCollection struct {
				ContributionCalendar struct{ TotalContributions int }
			}
			RepositoriesContributedTo struct{ TotalCount int }
		}
	}{}
	err = client.Do(query, variables, &resp)
	if err != nil {
		return nil, err
	}

	created, _ := time.Parse(time.RFC3339, resp.User.CreatedAt)
	age := time.Since(created)
	usr := User{
		Name:               username,
		Age:                age,
		Followers:          resp.User.Followers.TotalCount,
		Following:          resp.User.Following.TotalCount,
		TotalContributions: resp.User.ContributionsCollection.ContributionCalendar.TotalContributions,
		ReposContributed:   resp.User.RepositoriesContributedTo.TotalCount,
	}
	return &usr, nil
}

func getContributors(owner, repo string) (map[string]int, error) {
	opts := &api.ClientOptions{EnableCache: true}
	client, err := gh.RESTClient(opts)
	if err != nil {
		return nil, err
	}

	resp := []struct {
		Login         string
		Contributions int
	}{}
	err = client.Get(fmt.Sprintf("repos/%s/%s/contributors", owner, repo), &resp)
	if err != nil {
		return nil, err
	}

	contribs := make(map[string]int)

	for _, usr := range resp {
		fmt.Println(usr)
		contribs[usr.Login] = usr.Contributions
	}

	return contribs, nil
}

func issueSearchQuery(owner, repo, query string) ([]Issue, error) {
	opts := &api.ClientOptions{EnableCache: true}
	client, err := gh.GQLClient(opts)
	if err != nil {
		return nil, err
	}

	searchQuery := fmt.Sprintf("repo:%s/%s %s", owner, repo, query)
	gqlQuery := `query GetSpamIssues($query: String!, $after: String) {
search(query: $query, after: $after, type: ISSUE, first: 100) {
    pageInfo {
      hasNextPage
      endCursor
    }
    nodes {
      ... on Issue {
        author { login }
        title
		body
        number
        authorAssociation
      }
    }
  }
}`

	issues := []Issue{}
	variables := map[string]interface{}{"query": searchQuery}

	for {
		resp := struct {
			Search struct {
				PageInfo struct {
					HasNextPage bool
					EndCursor   string
				}
				Nodes []Issue
			}
		}{}

		err = client.Do(gqlQuery, variables, &resp)
		if err != nil {
			return nil, err
		}

		for _, issue := range resp.Search.Nodes {
			if issue.Title != "" {
				issues = append(issues, issue)
			}
		}

		if !resp.Search.PageInfo.HasNextPage {
			return issues, nil
		}
		variables["after"] = resp.Search.PageInfo.EndCursor
	}
}
