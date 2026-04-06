package suggest

import "sort"

// Levenshtein computes the edit distance between two strings using the
// Wagner–Fischer dynamic-programming algorithm.
func Levenshtein(a, b string) int {
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
			curr[j] = min(ins, del, sub)
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

type match struct {
	name     string
	distance int
}

// TopN returns up to maxResults candidates within maxDist edit distance of
// query. Results are sorted by distance (closest first), then alphabetically.
// Exact matches (distance 0) are excluded.
func TopN(query string, candidates []string, maxDist, maxResults int) []string {
	var matches []match
	for _, name := range candidates {
		d := Levenshtein(query, name)
		if d > 0 && d <= maxDist {
			matches = append(matches, match{name: name, distance: d})
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

// Best returns the single closest candidate to query within maxDist edit
// distance, or "" if nothing is close enough. When multiple candidates tie,
// the alphabetically first one wins.
func Best(query string, candidates []string, maxDist int) string {
	results := TopN(query, candidates, maxDist, 1)
	if len(results) == 0 {
		return ""
	}
	return results[0]
}
