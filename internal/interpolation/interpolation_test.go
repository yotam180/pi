package interpolation

import (
	"testing"
)

// --- OutputTracker ---

func TestOutputTracker_EmptyLast(t *testing.T) {
	var tr OutputTracker
	if got := tr.Last(); got != "" {
		t.Errorf("Last() on empty tracker = %q, want empty", got)
	}
}

func TestOutputTracker_RecordAndLast(t *testing.T) {
	var tr OutputTracker
	tr.Record("first")
	tr.Record("second")
	if got := tr.Last(); got != "second" {
		t.Errorf("Last() = %q, want %q", got, "second")
	}
}

func TestOutputTracker_RecordTrimsWhitespace(t *testing.T) {
	var tr OutputTracker
	tr.Record("  hello\n")
	if got := tr.Last(); got != "hello" {
		t.Errorf("Last() = %q, want %q", got, "hello")
	}
}

func TestOutputTracker_Get_Valid(t *testing.T) {
	var tr OutputTracker
	tr.Record("zero")
	tr.Record("one")
	tr.Record("two")

	for _, tc := range []struct {
		index int
		want  string
	}{
		{0, "zero"},
		{1, "one"},
		{2, "two"},
	} {
		got, ok := tr.Get(tc.index)
		if !ok {
			t.Errorf("Get(%d) returned not-ok", tc.index)
		}
		if got != tc.want {
			t.Errorf("Get(%d) = %q, want %q", tc.index, got, tc.want)
		}
	}
}

func TestOutputTracker_Get_OutOfRange(t *testing.T) {
	var tr OutputTracker
	tr.Record("only")

	for _, idx := range []int{-1, 1, 99} {
		val, ok := tr.Get(idx)
		if ok {
			t.Errorf("Get(%d) should not be ok", idx)
		}
		if val != "" {
			t.Errorf("Get(%d) = %q, want empty", idx, val)
		}
	}
}

func TestOutputTracker_Len(t *testing.T) {
	var tr OutputTracker
	if tr.Len() != 0 {
		t.Errorf("Len() = %d, want 0", tr.Len())
	}
	tr.Record("a")
	tr.Record("b")
	if tr.Len() != 2 {
		t.Errorf("Len() = %d, want 2", tr.Len())
	}
}

func TestOutputTracker_SnapshotRestore(t *testing.T) {
	var tr OutputTracker
	tr.Record("outer1")
	tr.Record("outer2")

	snap := tr.Snapshot()
	tr.Reset()
	tr.Record("inner")

	if tr.Len() != 1 {
		t.Errorf("after reset+record, Len() = %d, want 1", tr.Len())
	}
	if tr.Last() != "inner" {
		t.Errorf("after reset+record, Last() = %q, want %q", tr.Last(), "inner")
	}

	tr.Restore(snap)
	if tr.Len() != 2 {
		t.Errorf("after restore, Len() = %d, want 2", tr.Len())
	}
	if tr.Last() != "outer2" {
		t.Errorf("after restore, Last() = %q, want %q", tr.Last(), "outer2")
	}
}

func TestOutputTracker_SnapshotNil(t *testing.T) {
	var tr OutputTracker
	snap := tr.Snapshot()
	if snap != nil {
		t.Errorf("Snapshot() on empty tracker should be nil, got %v", snap)
	}

	tr.Record("something")
	tr.Restore(snap)
	if tr.Len() != 0 {
		t.Errorf("after restore(nil), Len() = %d, want 0", tr.Len())
	}
}

func TestOutputTracker_SnapshotIsACopy(t *testing.T) {
	var tr OutputTracker
	tr.Record("a")
	snap := tr.Snapshot()
	tr.Record("b")

	if len(snap) != 1 {
		t.Errorf("snapshot should have 1 entry, got %d", len(snap))
	}
	if snap[0] != "a" {
		t.Errorf("snapshot[0] = %q, want %q", snap[0], "a")
	}
}

func TestOutputTracker_Reset(t *testing.T) {
	var tr OutputTracker
	tr.Record("a")
	tr.Record("b")
	tr.Reset()
	if tr.Len() != 0 {
		t.Errorf("after Reset(), Len() = %d, want 0", tr.Len())
	}
	if tr.Last() != "" {
		t.Errorf("after Reset(), Last() = %q, want empty", tr.Last())
	}
}

// --- ResolveValue ---

func TestResolveValue_Literal(t *testing.T) {
	if got := ResolveValue("hello", nil, nil); got != "hello" {
		t.Errorf("ResolveValue(%q) = %q, want %q", "hello", got, "hello")
	}
}

func TestResolveValue_OutputsLast_Empty(t *testing.T) {
	var tr OutputTracker
	if got := ResolveValue("outputs.last", &tr, nil); got != "" {
		t.Errorf("ResolveValue(outputs.last) on empty = %q, want empty", got)
	}
}

func TestResolveValue_OutputsLast_NilTracker(t *testing.T) {
	if got := ResolveValue("outputs.last", nil, nil); got != "" {
		t.Errorf("ResolveValue(outputs.last) nil tracker = %q, want empty", got)
	}
}

func TestResolveValue_OutputsLast(t *testing.T) {
	var tr OutputTracker
	tr.Record("first")
	tr.Record("second")
	if got := ResolveValue("outputs.last", &tr, nil); got != "second" {
		t.Errorf("ResolveValue(outputs.last) = %q, want %q", got, "second")
	}
}

func TestResolveValue_OutputsIndexed(t *testing.T) {
	var tr OutputTracker
	tr.Record("zero")
	tr.Record("one")
	tr.Record("two")

	for _, tc := range []struct {
		input string
		want  string
	}{
		{"outputs.0", "zero"},
		{"outputs.1", "one"},
		{"outputs.2", "two"},
	} {
		if got := ResolveValue(tc.input, &tr, nil); got != tc.want {
			t.Errorf("ResolveValue(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestResolveValue_OutputsIndexed_OutOfRange(t *testing.T) {
	var tr OutputTracker
	tr.Record("only")
	if got := ResolveValue("outputs.99", &tr, nil); got != "outputs.99" {
		t.Errorf("ResolveValue(outputs.99) = %q, want passthrough", got)
	}
}

func TestResolveValue_OutputsIndexed_NilTracker(t *testing.T) {
	if got := ResolveValue("outputs.0", nil, nil); got != "outputs.0" {
		t.Errorf("ResolveValue(outputs.0) nil tracker = %q, want passthrough", got)
	}
}

func TestResolveValue_InputsFound(t *testing.T) {
	inputEnv := []string{"PI_IN_VERSION=22"}
	if got := ResolveValue("inputs.version", nil, inputEnv); got != "22" {
		t.Errorf("ResolveValue(inputs.version) = %q, want %q", got, "22")
	}
}

func TestResolveValue_InputsNotFound(t *testing.T) {
	inputEnv := []string{"PI_IN_VERSION=22"}
	if got := ResolveValue("inputs.unknown", nil, inputEnv); got != "inputs.unknown" {
		t.Errorf("ResolveValue(inputs.unknown) = %q, want passthrough", got)
	}
}

func TestResolveValue_InputsHyphenToUnderscore(t *testing.T) {
	inputEnv := []string{"PI_IN_MY_THING=abc"}
	if got := ResolveValue("inputs.my-thing", nil, inputEnv); got != "abc" {
		t.Errorf("ResolveValue(inputs.my-thing) = %q, want %q", got, "abc")
	}
}

func TestResolveValue_InputsEmptyEnv(t *testing.T) {
	if got := ResolveValue("inputs.version", nil, nil); got != "inputs.version" {
		t.Errorf("ResolveValue(inputs.version) nil env = %q, want passthrough", got)
	}
}

func TestResolveValue_InputsValueContainsEquals(t *testing.T) {
	inputEnv := []string{"PI_IN_QUERY=a=b=c"}
	if got := ResolveValue("inputs.query", nil, inputEnv); got != "a=b=c" {
		t.Errorf("ResolveValue(inputs.query) = %q, want %q", got, "a=b=c")
	}
}

// --- ResolveEnv ---

func TestResolveEnv_Nil(t *testing.T) {
	got := ResolveEnv(nil, nil, nil)
	if got != nil {
		t.Errorf("ResolveEnv(nil) = %v, want nil", got)
	}
}

func TestResolveEnv_Empty(t *testing.T) {
	got := ResolveEnv(map[string]string{}, nil, nil)
	if got == nil || len(got) != 0 {
		t.Errorf("ResolveEnv(empty) = %v, want empty map", got)
	}
}

func TestResolveEnv_NoChange(t *testing.T) {
	env := map[string]string{"KEY": "literal"}
	got := ResolveEnv(env, nil, nil)
	if got["KEY"] != "literal" {
		t.Errorf("ResolveEnv literal = %q, want %q", got["KEY"], "literal")
	}
}

func TestResolveEnv_WithOutputsLast(t *testing.T) {
	var tr OutputTracker
	tr.Record("first")
	tr.Record("second")

	env := map[string]string{"VER": "outputs.last"}
	got := ResolveEnv(env, &tr, nil)
	if got["VER"] != "second" {
		t.Errorf("ResolveEnv outputs.last = %q, want %q", got["VER"], "second")
	}
}

func TestResolveEnv_WithOutputsIndexed(t *testing.T) {
	var tr OutputTracker
	tr.Record("zero")
	tr.Record("one")

	env := map[string]string{"A": "outputs.0", "B": "outputs.1"}
	got := ResolveEnv(env, &tr, nil)
	if got["A"] != "zero" {
		t.Errorf("ResolveEnv outputs.0 = %q, want %q", got["A"], "zero")
	}
	if got["B"] != "one" {
		t.Errorf("ResolveEnv outputs.1 = %q, want %q", got["B"], "one")
	}
}

func TestResolveEnv_WithInputs(t *testing.T) {
	inputEnv := []string{"PI_IN_NAME=world"}
	env := map[string]string{"GREETING": "inputs.name"}
	got := ResolveEnv(env, nil, inputEnv)
	if got["GREETING"] != "world" {
		t.Errorf("ResolveEnv inputs.name = %q, want %q", got["GREETING"], "world")
	}
}

func TestResolveEnv_MixedLiteralsAndRefs(t *testing.T) {
	var tr OutputTracker
	tr.Record("captured")

	inputEnv := []string{"PI_IN_VER=3"}
	env := map[string]string{
		"A": "outputs.last",
		"B": "inputs.ver",
		"C": "literal-value",
	}
	got := ResolveEnv(env, &tr, inputEnv)
	if got["A"] != "captured" {
		t.Errorf("A = %q, want %q", got["A"], "captured")
	}
	if got["B"] != "3" {
		t.Errorf("B = %q, want %q", got["B"], "3")
	}
	if got["C"] != "literal-value" {
		t.Errorf("C = %q, want %q", got["C"], "literal-value")
	}
}

func TestResolveEnv_ReturnsOriginalWhenUnchanged(t *testing.T) {
	env := map[string]string{"KEY": "value"}
	got := ResolveEnv(env, nil, nil)
	// When nothing changes, the original map should be returned
	if &got == nil {
		t.Error("expected non-nil result")
	}
	if got["KEY"] != "value" {
		t.Errorf("got[KEY] = %q, want %q", got["KEY"], "value")
	}
}

// --- ResolveWith ---

func TestResolveWith_Nil(t *testing.T) {
	got := ResolveWith(nil, nil, nil)
	if got != nil {
		t.Errorf("ResolveWith(nil) = %v, want nil", got)
	}
}

func TestResolveWith_Empty(t *testing.T) {
	got := ResolveWith(map[string]string{}, nil, nil)
	if got == nil || len(got) != 0 {
		t.Errorf("ResolveWith(empty) = %v, want empty map", got)
	}
}

func TestResolveWith_Literal(t *testing.T) {
	got := ResolveWith(map[string]string{"key": "literal-value"}, nil, nil)
	if got["key"] != "literal-value" {
		t.Errorf("ResolveWith literal = %q, want %q", got["key"], "literal-value")
	}
}

func TestResolveWith_OutputsLast(t *testing.T) {
	var tr OutputTracker
	tr.Record("hello")
	tr.Record("world")
	got := ResolveWith(map[string]string{"key": "outputs.last"}, &tr, nil)
	if got["key"] != "world" {
		t.Errorf("ResolveWith outputs.last = %q, want %q", got["key"], "world")
	}
}

func TestResolveWith_Inputs(t *testing.T) {
	inputEnv := []string{"PI_IN_VER=42"}
	got := ResolveWith(map[string]string{"ver": "inputs.ver"}, nil, inputEnv)
	if got["ver"] != "42" {
		t.Errorf("ResolveWith inputs = %q, want %q", got["ver"], "42")
	}
}

func TestResolveWith_MultipleKeys(t *testing.T) {
	var tr OutputTracker
	tr.Record("v1")
	tr.Record("v2")

	with := map[string]string{
		"a": "outputs.0",
		"b": "outputs.last",
		"c": "literal",
	}
	got := ResolveWith(with, &tr, nil)
	if got["a"] != "v1" {
		t.Errorf("a = %q, want %q", got["a"], "v1")
	}
	if got["b"] != "v2" {
		t.Errorf("b = %q, want %q", got["b"], "v2")
	}
	if got["c"] != "literal" {
		t.Errorf("c = %q, want %q", got["c"], "literal")
	}
}
