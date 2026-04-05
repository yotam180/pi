package discovery

import "sort"

// levenshtein computes the edit distance between two strings using the
// Wagner–Fischer dynamic-programming algorithm.
func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	prev := make([]int, lb+1)
	for j := range prev {
		prev[j] = j
	}

	curr := make([]int, lb+1)
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			ins := curr[j-1] + 1
			del := prev[j] + 1
			sub := prev[j-1] + cost
			curr[j] = min3(ins, del, sub)
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

type suggestion struct {
	name     string
	distance int
}

// suggestNames returns up to maxResults automation names from candidates that
// are within a reasonable edit distance of query. Results are sorted by
// distance (closest first), then alphabetically for ties.
func suggestNames(query string, candidates []string, maxResults int) []string {
	maxDist := len(query) * 30 / 100
	if maxDist < 3 {
		maxDist = 3
	}

	var matches []suggestion
	for _, name := range candidates {
		d := levenshtein(query, name)
		if d > 0 && d <= maxDist {
			matches = append(matches, suggestion{name: name, distance: d})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].distance != matches[j].distance {
			return matches[i].distance < matches[j].distance
		}
		return matches[i].name < matches[j].name
	})

	if len(matches) > maxResults {
		matches = matches[:maxResults]
	}

	result := make([]string, len(matches))
	for i, m := range matches {
		result[i] = m.name
	}
	return result
}
