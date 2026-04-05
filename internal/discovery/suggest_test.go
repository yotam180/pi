package discovery

import (
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
)

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"abc", "", 3},
		{"", "abc", 3},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"kitten", "sitting", 3},
		{"docker/up", "dokcer/up", 2},
		{"docker/up", "docker/down", 4},
		{"build", "biuld", 2},
		{"a", "b", 1},
	}

	for _, tt := range tests {
		got := levenshtein(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("levenshtein(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestLevenshtein_Symmetry(t *testing.T) {
	pairs := [][2]string{
		{"docker/up", "dokcer/up"},
		{"build", "guild"},
		{"test", "best"},
	}
	for _, pair := range pairs {
		ab := levenshtein(pair[0], pair[1])
		ba := levenshtein(pair[1], pair[0])
		if ab != ba {
			t.Errorf("levenshtein(%q, %q) = %d, but levenshtein(%q, %q) = %d", pair[0], pair[1], ab, pair[1], pair[0], ba)
		}
	}
}

func TestSuggestNames_CloseTypo(t *testing.T) {
	candidates := []string{"build", "deploy", "docker/up", "docker/down", "test"}
	got := suggestNames("dokcer/up", candidates, 3)
	if len(got) == 0 {
		t.Fatal("expected suggestions for 'dokcer/up'")
	}
	if got[0] != "docker/up" {
		t.Errorf("first suggestion should be 'docker/up', got %q", got[0])
	}
}

func TestSuggestNames_ExactMatch_NoSuggestion(t *testing.T) {
	candidates := []string{"build", "docker/up", "test"}
	got := suggestNames("docker/up", candidates, 3)
	if len(got) != 0 {
		t.Errorf("expected no suggestions for exact match, got %v", got)
	}
}

func TestSuggestNames_NoCloseMatches(t *testing.T) {
	candidates := []string{"build", "deploy", "test"}
	got := suggestNames("zzzzzzzzzzzzz", candidates, 3)
	if len(got) != 0 {
		t.Errorf("expected no suggestions for distant string, got %v", got)
	}
}

func TestSuggestNames_MaxResults(t *testing.T) {
	candidates := []string{"aaa", "aab", "aac", "aad", "aae"}
	got := suggestNames("aax", candidates, 2)
	if len(got) > 2 {
		t.Errorf("expected at most 2 suggestions, got %d: %v", len(got), got)
	}
}

func TestSuggestNames_SortedByDistance(t *testing.T) {
	candidates := []string{"docker/down", "docker/up", "docker/logs"}
	got := suggestNames("docker/pu", candidates, 3)
	if len(got) == 0 {
		t.Fatal("expected suggestions")
	}
	if got[0] != "docker/up" {
		t.Errorf("closest match should be 'docker/up' (distance 1), got %q", got[0])
	}
}

func TestSuggestNames_AlphabeticalTiebreaker(t *testing.T) {
	candidates := []string{"bbb", "aaa", "ccc"}
	got := suggestNames("xxx", candidates, 3)
	for i := 1; i < len(got); i++ {
		di := levenshtein("xxx", got[i-1])
		dj := levenshtein("xxx", got[i])
		if di == dj && got[i-1] > got[i] {
			t.Errorf("suggestions with equal distance should be alphabetical: %v", got)
		}
	}
}

func TestSuggestNames_EmptyCandidates(t *testing.T) {
	got := suggestNames("docker/up", nil, 3)
	if len(got) != 0 {
		t.Errorf("expected no suggestions for empty candidates, got %v", got)
	}
}

func TestSuggestNames_EmptyQuery(t *testing.T) {
	candidates := []string{"a", "ab", "abc"}
	got := suggestNames("", candidates, 3)
	for _, s := range got {
		if levenshtein("", s) > 3 {
			t.Errorf("suggestion %q too far from empty query", s)
		}
	}
}

func TestFindLocal_IncludesDidYouMean(t *testing.T) {
	autos := map[string]*automation.Automation{
		"docker/up":   {Name: "docker/up", Description: "Start containers"},
		"docker/down": {Name: "docker/down", Description: "Stop containers"},
		"build":       {Name: "build"},
	}
	names := []string{"build", "docker/down", "docker/up"}
	r := NewResult(autos, names)

	_, err := r.Find("dokcer/up")
	if err == nil {
		t.Fatal("expected error for misspelled name")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "Did you mean?") {
		t.Errorf("error should contain 'Did you mean?': %v", err)
	}
	if !strings.Contains(errStr, "docker/up") {
		t.Errorf("error should suggest 'docker/up': %v", err)
	}
}

func TestFindLocal_NoSuggestions_WhenDistant(t *testing.T) {
	autos := map[string]*automation.Automation{
		"docker/up": {Name: "docker/up"},
		"build":     {Name: "build"},
	}
	names := []string{"build", "docker/up"}
	r := NewResult(autos, names)

	_, err := r.Find("zzzzzzzzzzzzzzzzzzzzz")
	if err == nil {
		t.Fatal("expected error")
	}

	errStr := err.Error()
	if strings.Contains(errStr, "Did you mean?") {
		t.Errorf("should NOT contain 'Did you mean?' for distant names: %v", err)
	}
	if !strings.Contains(errStr, "Available automations:") {
		t.Errorf("should still list available automations: %v", err)
	}
}

func TestFindBuiltin_IncludesDidYouMean(t *testing.T) {
	builtins := map[string]*automation.Automation{
		"install-python": {Name: "install-python"},
		"install-node":   {Name: "install-node"},
		"docker/up":      {Name: "docker/up"},
	}
	builtinNames := []string{"docker/up", "install-node", "install-python"}
	builtinResult := NewResult(builtins, builtinNames)

	localAutos := map[string]*automation.Automation{}
	r := NewResult(localAutos, nil)
	r.MergeBuiltins(builtinResult)

	_, err := r.Find("pi:install-pyhton")
	if err == nil {
		t.Fatal("expected error for misspelled builtin name")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "Did you mean?") {
		t.Errorf("error should contain 'Did you mean?': %v", err)
	}
	if !strings.Contains(errStr, "pi:install-python") {
		t.Errorf("error should suggest 'pi:install-python': %v", err)
	}
}
