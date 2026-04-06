package discovery

import "github.com/vyper-tooling/pi/internal/suggest"

// suggestNames returns up to maxResults automation names from candidates that
// are within a reasonable edit distance of query. Results are sorted by
// distance (closest first), then alphabetically for ties.
func suggestNames(query string, candidates []string, maxResults int) []string {
	maxDist := len(query) * 30 / 100
	if maxDist < 3 {
		maxDist = 3
	}
	return suggest.TopN(query, candidates, maxDist, maxResults)
}
