package spam

import "github.com/ktr0731/go-fuzzyfinder/scoring"

func simScores(query string, docs []string) []int {
	sims := []int{}

	for _, doc := range docs {
		var sim int
		if len(query) >= len(doc) {
			sim, _ = scoring.Calculate(query, doc)
		} else {
			sim, _ = scoring.Calculate(doc, query)
		}

		sims = append(sims, sim)
	}
	return sims
}

func MaxSimScore(query string, docs []string) int {
	if query == "" {
		return 0
	}

	sims := simScores(query, docs)
	if len(sims) < 1 {
		return 0
	}
	max := sims[0]
	for _, sim := range sims {
		if sim > max {
			max = sim
		}
	}
	return max
}
