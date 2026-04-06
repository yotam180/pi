package suggest

import "testing"

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
		got := Levenshtein(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("Levenshtein(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
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
		ab := Levenshtein(pair[0], pair[1])
		ba := Levenshtein(pair[1], pair[0])
		if ab != ba {
			t.Errorf("Levenshtein(%q, %q) = %d, but Levenshtein(%q, %q) = %d",
				pair[0], pair[1], ab, pair[1], pair[0], ba)
		}
	}
}

func TestTopN_CloseTypo(t *testing.T) {
	candidates := []string{"build", "deploy", "docker/up", "docker/down", "test"}
	got := TopN("dokcer/up", candidates, 3, 3)
	if len(got) == 0 {
		t.Fatal("expected suggestions for 'dokcer/up'")
	}
	if got[0] != "docker/up" {
		t.Errorf("first suggestion should be 'docker/up', got %q", got[0])
	}
}

func TestTopN_ExactMatch_Excluded(t *testing.T) {
	candidates := []string{"build", "docker/up", "test"}
	got := TopN("docker/up", candidates, 3, 3)
	if len(got) != 0 {
		t.Errorf("expected no suggestions for exact match, got %v", got)
	}
}

func TestTopN_NoCloseMatches(t *testing.T) {
	candidates := []string{"build", "deploy", "test"}
	got := TopN("zzzzzzzzzzzzz", candidates, 3, 3)
	if len(got) != 0 {
		t.Errorf("expected no suggestions for distant string, got %v", got)
	}
}

func TestTopN_MaxResults(t *testing.T) {
	candidates := []string{"aaa", "aab", "aac", "aad", "aae"}
	got := TopN("aax", candidates, 3, 2)
	if len(got) > 2 {
		t.Errorf("expected at most 2 suggestions, got %d: %v", len(got), got)
	}
}

func TestTopN_SortedByDistance(t *testing.T) {
	candidates := []string{"docker/down", "docker/up", "docker/logs"}
	got := TopN("docker/pu", candidates, 3, 3)
	if len(got) == 0 {
		t.Fatal("expected suggestions")
	}
	if got[0] != "docker/up" {
		t.Errorf("closest match should be 'docker/up' (distance 1), got %q", got[0])
	}
}

func TestTopN_AlphabeticalTiebreaker(t *testing.T) {
	candidates := []string{"bbb", "aaa", "ccc"}
	got := TopN("xxx", candidates, 3, 3)
	for i := 1; i < len(got); i++ {
		di := Levenshtein("xxx", got[i-1])
		dj := Levenshtein("xxx", got[i])
		if di == dj && got[i-1] > got[i] {
			t.Errorf("suggestions with equal distance should be alphabetical: %v", got)
		}
	}
}

func TestTopN_EmptyCandidates(t *testing.T) {
	got := TopN("docker/up", nil, 3, 3)
	if len(got) != 0 {
		t.Errorf("expected no suggestions for empty candidates, got %v", got)
	}
}

func TestTopN_EmptyQuery(t *testing.T) {
	candidates := []string{"a", "ab", "abc"}
	got := TopN("", candidates, 3, 3)
	for _, s := range got {
		if Levenshtein("", s) > 3 {
			t.Errorf("suggestion %q too far from empty query", s)
		}
	}
}

func TestBest_CloseMatch(t *testing.T) {
	candidates := []string{"bash", "python", "typescript", "run"}
	got := Best("bsh", candidates, 2)
	if got != "bash" {
		t.Errorf("Best(%q) = %q, want %q", "bsh", got, "bash")
	}
}

func TestBest_NoMatch(t *testing.T) {
	candidates := []string{"bash", "python", "typescript"}
	got := Best("zzzzzzz", candidates, 2)
	if got != "" {
		t.Errorf("Best(%q) = %q, want empty", "zzzzzzz", got)
	}
}

func TestBest_ExactMatch_Excluded(t *testing.T) {
	candidates := []string{"bash", "python"}
	got := Best("bash", candidates, 2)
	if got != "" {
		t.Errorf("Best(%q) = %q, want empty (exact match excluded)", "bash", got)
	}
}

func TestBest_MultipleCandidates(t *testing.T) {
	candidates := []string{"steps", "step", "stesp"}
	got := Best("stpes", candidates, 2)
	if got == "" {
		t.Fatal("expected a suggestion")
	}
}

func TestBest_ShortField(t *testing.T) {
	candidates := []string{"if", "env"}
	got := Best("fi", candidates, 2)
	if got == "" {
		t.Error("expected a suggestion for short field")
	}
}

func TestBest_AlphabeticalTiebreaker(t *testing.T) {
	candidates := []string{"bbb", "aaa"}
	got := Best("aab", candidates, 2)
	if got != "aaa" {
		t.Errorf("Best(%q) = %q, want %q (alphabetical tiebreaker)", "aab", got, "aaa")
	}
}
