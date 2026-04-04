package conditions

import (
	"strings"
	"testing"
)

// --- Eval tests ---

func TestEvalEmptyExpression(t *testing.T) {
	for _, expr := range []string{"", "  ", "\t\n"} {
		result, err := Eval(expr, nil)
		if err != nil {
			t.Errorf("Eval(%q): unexpected error: %v", expr, err)
		}
		if !result {
			t.Errorf("Eval(%q) = false, want true", expr)
		}
	}
}

func TestEvalBareIdentifier(t *testing.T) {
	preds := map[string]bool{
		"os.macos": true,
		"os.linux": false,
	}

	result, err := Eval("os.macos", preds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected true for os.macos")
	}

	result, err = Eval("os.linux", preds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result {
		t.Error("expected false for os.linux")
	}
}

func TestEvalDottedIdentifiers(t *testing.T) {
	preds := map[string]bool{
		"os.arch.arm64":  true,
		"command.docker": false,
		"env.CI":         true,
		"shell.zsh":      true,
	}

	tests := []struct {
		expr string
		want bool
	}{
		{"os.arch.arm64", true},
		{"command.docker", false},
		{"env.CI", true},
		{"shell.zsh", true},
	}

	for _, tt := range tests {
		result, err := Eval(tt.expr, preds)
		if err != nil {
			t.Errorf("Eval(%q): unexpected error: %v", tt.expr, err)
			continue
		}
		if result != tt.want {
			t.Errorf("Eval(%q) = %v, want %v", tt.expr, result, tt.want)
		}
	}
}

func TestEvalNotOperator(t *testing.T) {
	preds := map[string]bool{
		"env.CI":         true,
		"command.docker": false,
	}

	tests := []struct {
		expr string
		want bool
	}{
		{"not env.CI", false},
		{"not command.docker", true},
	}

	for _, tt := range tests {
		result, err := Eval(tt.expr, preds)
		if err != nil {
			t.Errorf("Eval(%q): unexpected error: %v", tt.expr, err)
			continue
		}
		if result != tt.want {
			t.Errorf("Eval(%q) = %v, want %v", tt.expr, result, tt.want)
		}
	}
}

func TestEvalDoubleNot(t *testing.T) {
	preds := map[string]bool{
		"env.CI": true,
	}

	result, err := Eval("not not env.CI", preds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected true for not not env.CI (where env.CI is true)")
	}
}

func TestEvalAndOperator(t *testing.T) {
	preds := map[string]bool{
		"os.macos":     true,
		"command.brew": true,
		"env.CI":       false,
	}

	tests := []struct {
		expr string
		want bool
	}{
		{"os.macos and command.brew", true},
		{"os.macos and env.CI", false},
		{"env.CI and os.macos", false},
	}

	for _, tt := range tests {
		result, err := Eval(tt.expr, preds)
		if err != nil {
			t.Errorf("Eval(%q): unexpected error: %v", tt.expr, err)
			continue
		}
		if result != tt.want {
			t.Errorf("Eval(%q) = %v, want %v", tt.expr, result, tt.want)
		}
	}
}

func TestEvalOrOperator(t *testing.T) {
	preds := map[string]bool{
		"os.macos": false,
		"os.linux": true,
		"env.CI":   false,
	}

	tests := []struct {
		expr string
		want bool
	}{
		{"os.macos or os.linux", true},
		{"os.macos or env.CI", false},
	}

	for _, tt := range tests {
		result, err := Eval(tt.expr, preds)
		if err != nil {
			t.Errorf("Eval(%q): unexpected error: %v", tt.expr, err)
			continue
		}
		if result != tt.want {
			t.Errorf("Eval(%q) = %v, want %v", tt.expr, result, tt.want)
		}
	}
}

func TestEvalPrecedence(t *testing.T) {
	// "and" binds tighter than "or"
	// a or b and c  →  a or (b and c)
	preds := map[string]bool{
		"a": true,
		"b": false,
		"c": true,
	}

	// true or (false and true) → true
	result, err := Eval("a or b and c", preds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected true for 'a or b and c' with a=true, b=false, c=true")
	}

	// (false and true) or false  →  false
	preds2 := map[string]bool{
		"a": false,
		"b": true,
		"c": false,
	}
	result, err = Eval("a and b or c", preds2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result {
		t.Error("expected false for 'a and b or c' with a=false, b=true, c=false")
	}
}

func TestEvalParentheses(t *testing.T) {
	preds := map[string]bool{
		"os.macos":       true,
		"os.linux":       false,
		"command.docker": true,
	}

	// (true or false) and true → true
	result, err := Eval("(os.macos or os.linux) and command.docker", preds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected true")
	}
}

func TestEvalNestedParentheses(t *testing.T) {
	preds := map[string]bool{
		"a": true,
		"b": false,
		"c": true,
	}

	// ((true) and (not false)) or false → true
	result, err := Eval("((a) and (not b)) or c", preds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected true")
	}
}

func TestEvalNotWithAnd(t *testing.T) {
	preds := map[string]bool{
		"os.macos":     true,
		"command.brew": false,
	}

	// true and not false → true
	result, err := Eval("os.macos and not command.brew", preds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected true for os.macos and not command.brew")
	}
}

func TestEvalFunctionCall(t *testing.T) {
	preds := map[string]bool{
		"file.exists(\".env\")":      true,
		"dir.exists(\"data/logs\")":  false,
	}

	tests := []struct {
		expr string
		want bool
	}{
		{`file.exists(".env")`, true},
		{`dir.exists("data/logs")`, false},
	}

	for _, tt := range tests {
		result, err := Eval(tt.expr, preds)
		if err != nil {
			t.Errorf("Eval(%q): unexpected error: %v", tt.expr, err)
			continue
		}
		if result != tt.want {
			t.Errorf("Eval(%q) = %v, want %v", tt.expr, result, tt.want)
		}
	}
}

func TestEvalFunctionCallWithOperators(t *testing.T) {
	preds := map[string]bool{
		"file.exists(\".env\")": true,
		"env.CI":                false,
	}

	result, err := Eval(`file.exists(".env") and not env.CI`, preds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected true")
	}
}

func TestEvalComplexExpression(t *testing.T) {
	preds := map[string]bool{
		"os.macos":              true,
		"os.linux":              false,
		"command.docker":        true,
		"env.DOCKER_HOST":       false,
		"file.exists(\".env\")": true,
		"env.CI":                false,
	}

	tests := []struct {
		expr string
		want bool
	}{
		{"(os.macos or os.linux) and command.docker", true},
		{"env.DOCKER_HOST or command.docker", true},
		{`file.exists(".env") and not env.CI`, true},
		{"os.linux and command.docker", false},
	}

	for _, tt := range tests {
		result, err := Eval(tt.expr, preds)
		if err != nil {
			t.Errorf("Eval(%q): unexpected error: %v", tt.expr, err)
			continue
		}
		if result != tt.want {
			t.Errorf("Eval(%q) = %v, want %v", tt.expr, result, tt.want)
		}
	}
}

func TestEvalMultipleAndOr(t *testing.T) {
	preds := map[string]bool{
		"a": true,
		"b": true,
		"c": true,
		"d": false,
	}

	// a and b and c → true
	result, err := Eval("a and b and c", preds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected true for a and b and c")
	}

	// a or b or d → true
	result, err = Eval("a or b or d", preds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected true for a or b or d")
	}

	// a and b and d → false
	result, err = Eval("a and b and d", preds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result {
		t.Error("expected false for a and b and d")
	}
}

// --- Error cases ---

func TestEvalUnknownPredicate(t *testing.T) {
	preds := map[string]bool{"os.macos": true}

	_, err := Eval("os.macos and command.missing", preds)
	if err == nil {
		t.Fatal("expected error for unknown predicate")
	}
	if !strings.Contains(err.Error(), "command.missing") {
		t.Errorf("error should mention the unknown predicate, got: %v", err)
	}
}

func TestEvalUnknownFuncCallPredicate(t *testing.T) {
	preds := map[string]bool{}

	_, err := Eval(`file.exists("missing.txt")`, preds)
	if err == nil {
		t.Fatal("expected error for unknown function predicate")
	}
	if !strings.Contains(err.Error(), "file.exists") || !strings.Contains(err.Error(), "missing.txt") {
		t.Errorf("error should mention the function predicate, got: %v", err)
	}
}

func TestEvalMalformedExpressions(t *testing.T) {
	tests := []struct {
		expr    string
		wantErr string
	}{
		{"and os.macos", "unexpected token"},
		{"os.macos and", "unexpected end"},
		{"(os.macos", "expected ')'"},
		{"os.macos)", "unexpected token"},
		{"os.macos or or os.linux", "unexpected token"},
		{`file.exists(123)`, "unexpected character"},
		{`file.exists("foo"`, "expected ')'"},
		{`"just a string"`, "unexpected token"},
		{"@#$", "unexpected character"},
	}

	for _, tt := range tests {
		_, err := Eval(tt.expr, map[string]bool{})
		if err == nil {
			t.Errorf("Eval(%q): expected error containing %q, got nil", tt.expr, tt.wantErr)
			continue
		}
		if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.wantErr)) {
			t.Errorf("Eval(%q): error %q should contain %q", tt.expr, err.Error(), tt.wantErr)
		}
	}
}

func TestEvalUnterminatedString(t *testing.T) {
	_, err := Eval(`file.exists("unterminated`, map[string]bool{})
	if err == nil {
		t.Fatal("expected error for unterminated string")
	}
	if !strings.Contains(err.Error(), "unterminated string") {
		t.Errorf("error should mention unterminated string, got: %v", err)
	}
}

// --- Predicates tests ---

func TestPredicatesEmpty(t *testing.T) {
	preds, err := Predicates("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(preds) != 0 {
		t.Errorf("expected empty slice, got %v", preds)
	}
}

func TestPredicatesSingleIdent(t *testing.T) {
	preds, err := Predicates("os.macos")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(preds) != 1 || preds[0] != "os.macos" {
		t.Errorf("expected [os.macos], got %v", preds)
	}
}

func TestPredicatesComplex(t *testing.T) {
	preds, err := Predicates(`(os.macos or os.linux) and command.docker and file.exists(".env")`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"os.macos", "os.linux", "command.docker", `file.exists(".env")`}
	if len(preds) != len(expected) {
		t.Fatalf("expected %d predicates, got %d: %v", len(expected), len(preds), preds)
	}
	for i, exp := range expected {
		if preds[i] != exp {
			t.Errorf("predicate[%d] = %q, want %q", i, preds[i], exp)
		}
	}
}

func TestPredicatesDeduplication(t *testing.T) {
	preds, err := Predicates("os.macos and os.macos or os.macos")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(preds) != 1 {
		t.Errorf("expected 1 unique predicate, got %d: %v", len(preds), preds)
	}
}

func TestPredicatesNotOperator(t *testing.T) {
	preds, err := Predicates("not env.CI")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(preds) != 1 || preds[0] != "env.CI" {
		t.Errorf("expected [env.CI], got %v", preds)
	}
}

func TestPredicatesMalformed(t *testing.T) {
	_, err := Predicates("and or")
	if err == nil {
		t.Fatal("expected error for malformed expression")
	}
}

func TestPredicatesFunctionCalls(t *testing.T) {
	preds, err := Predicates(`file.exists(".env") and dir.exists("data")`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{`file.exists(".env")`, `dir.exists("data")`}
	if len(preds) != len(expected) {
		t.Fatalf("expected %d predicates, got %d: %v", len(expected), len(preds), preds)
	}
	for i, exp := range expected {
		if preds[i] != exp {
			t.Errorf("predicate[%d] = %q, want %q", i, preds[i], exp)
		}
	}
}

func TestPredicatesWhitespace(t *testing.T) {
	preds, err := Predicates("   os.macos   and   command.brew   ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(preds) != 2 {
		t.Fatalf("expected 2 predicates, got %d: %v", len(preds), preds)
	}
	if preds[0] != "os.macos" || preds[1] != "command.brew" {
		t.Errorf("got %v", preds)
	}
}

// --- Lexer edge case tests ---

func TestLexerEmptyInput(t *testing.T) {
	tokens, err := lex("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 1 || tokens[0].typ != tokenEOF {
		t.Errorf("expected single EOF token, got %d tokens", len(tokens))
	}
}

func TestLexerPositions(t *testing.T) {
	tokens, err := lex("a and b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// a=0, and=2, b=6, EOF=7
	if tokens[0].pos != 0 {
		t.Errorf("token 'a' pos = %d, want 0", tokens[0].pos)
	}
	if tokens[1].pos != 2 {
		t.Errorf("token 'and' pos = %d, want 2", tokens[1].pos)
	}
	if tokens[2].pos != 6 {
		t.Errorf("token 'b' pos = %d, want 6", tokens[2].pos)
	}
}

// --- Table-driven spec examples ---

func TestEvalSpecExamples(t *testing.T) {
	preds := map[string]bool{
		"os.macos":              true,
		"os.linux":              false,
		"env.CI":                false,
		"command.brew":          true,
		"command.docker":        true,
		"env.DOCKER_HOST":       true,
		"file.exists(\".env\")": true,
	}

	tests := []struct {
		expr string
		want bool
	}{
		{"os.macos", true},
		{"not env.CI", true},
		{"os.macos and command.brew", true},
		{"(os.macos or os.linux) and command.docker", true},
		{"os.macos and not command.brew", false},
		{"env.DOCKER_HOST or command.docker", true},
		{`file.exists(".env") and not env.CI`, true},
	}

	for _, tt := range tests {
		result, err := Eval(tt.expr, preds)
		if err != nil {
			t.Errorf("Eval(%q): unexpected error: %v", tt.expr, err)
			continue
		}
		if result != tt.want {
			t.Errorf("Eval(%q) = %v, want %v", tt.expr, result, tt.want)
		}
	}
}
