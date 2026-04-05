package integration

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestVersion_Flag(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "--version")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	trimmed := strings.TrimSpace(out)
	if !strings.HasPrefix(trimmed, "pi ") {
		t.Errorf("expected output starting with 'pi ', got %q", trimmed)
	}
	if len(strings.TrimPrefix(trimmed, "pi ")) == 0 {
		t.Error("expected non-empty version string after 'pi '")
	}
}

func TestVersion_Subcommand(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "version")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	trimmed := strings.TrimSpace(out)
	if !strings.HasPrefix(trimmed, "pi ") {
		t.Errorf("expected output starting with 'pi ', got %q", trimmed)
	}
}

func TestVersion_FlagAndSubcommandMatch(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	flagOut, flagCode := runPi(t, dir, "--version")
	subOut, subCode := runPi(t, dir, "version")
	if flagCode != 0 || subCode != 0 {
		t.Fatalf("expected exit 0 for both, got flag=%d sub=%d", flagCode, subCode)
	}
	if strings.TrimSpace(flagOut) != strings.TrimSpace(subOut) {
		t.Errorf("--version and version subcommand differ: %q vs %q", flagOut, subOut)
	}
}
