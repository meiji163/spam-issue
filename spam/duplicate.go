package spam

import "github.com/ktr0731/go-fuzzyfinder/scoring"

func Similarities(query string, docs []string) []int {
	sims := []int{}

	for _, doc := range docs {
		var sim int
		if len(query) < len(doc) {
			sim, _ = scoring.Calculate(query, doc)
		} else {
			sim, _ = scoring.Calculate(doc, query)
		}

		sims = append(sims, sim)
	}
	return sims
}
