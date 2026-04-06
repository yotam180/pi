package conditions

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func testEnv(goos, goarch string, envVars map[string]string, commands map[string]bool) *RuntimeEnv {
	return &RuntimeEnv{
		GOOS:   goos,
		GOARCH: goarch,
		Getenv: func(key string) string {
			return envVars[key]
		},
		LookPath: func(name string) (string, error) {
			if commands[name] {
				return "/usr/bin/" + name, nil
			}
			return "", fmt.Errorf("not found")
		},
		Stat: os.Stat,
	}
}

func TestOSPredicates(t *testing.T) {
	tests := []struct {
		predicate string
		goos      string
		want      bool
	}{
		{"os.macos", "darwin", true},
		{"os.macos", "linux", false},
		{"os.macos", "windows", false},
		{"os.linux", "linux", true},
		{"os.linux", "darwin", false},
		{"os.linux", "windows", false},
		{"os.windows", "windows", true},
		{"os.windows", "darwin", false},
		{"os.windows", "linux", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_on_%s", tt.predicate, tt.goos), func(t *testing.T) {
			env := testEnv(tt.goos, "amd64", nil, nil)
			result, err := ResolvePredicatesWithEnv([]string{tt.predicate}, "/tmp", env)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result[tt.predicate] != tt.want {
				t.Errorf("got %v, want %v", result[tt.predicate], tt.want)
			}
		})
	}
}

func TestArchPredicates(t *testing.T) {
	tests := []struct {
		predicate string
		goarch    string
		want      bool
	}{
		{"os.arch.arm64", "arm64", true},
		{"os.arch.arm64", "amd64", false},
		{"os.arch.amd64", "amd64", true},
		{"os.arch.amd64", "arm64", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_on_%s", tt.predicate, tt.goarch), func(t *testing.T) {
			env := testEnv("linux", tt.goarch, nil, nil)
			result, err := ResolvePredicatesWithEnv([]string{tt.predicate}, "/tmp", env)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result[tt.predicate] != tt.want {
				t.Errorf("got %v, want %v", result[tt.predicate], tt.want)
			}
		})
	}
}

func TestShellPredicates(t *testing.T) {
	tests := []struct {
		predicate string
		shell     string
		want      bool
	}{
		{"shell.zsh", "/bin/zsh", true},
		{"shell.zsh", "/usr/bin/zsh", true},
		{"shell.zsh", "/bin/bash", false},
		{"shell.zsh", "", false},
		{"shell.bash", "/bin/bash", true},
		{"shell.bash", "/usr/bin/bash", true},
		{"shell.bash", "/bin/zsh", false},
		{"shell.bash", "", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_with_%s", tt.predicate, tt.shell), func(t *testing.T) {
			envVars := map[string]string{"SHELL": tt.shell}
			env := testEnv("linux", "amd64", envVars, nil)
			result, err := ResolvePredicatesWithEnv([]string{tt.predicate}, "/tmp", env)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result[tt.predicate] != tt.want {
				t.Errorf("got %v, want %v", result[tt.predicate], tt.want)
			}
		})
	}
}

func TestEnvPredicates(t *testing.T) {
	t.Run("env_var_set", func(t *testing.T) {
		envVars := map[string]string{"MY_VAR": "hello"}
		env := testEnv("linux", "amd64", envVars, nil)
		result, err := ResolvePredicatesWithEnv([]string{"env.MY_VAR"}, "/tmp", env)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result["env.MY_VAR"] {
			t.Error("expected true for set env var")
		}
	})

	t.Run("env_var_not_set", func(t *testing.T) {
		env := testEnv("linux", "amd64", nil, nil)
		result, err := ResolvePredicatesWithEnv([]string{"env.MISSING_VAR"}, "/tmp", env)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["env.MISSING_VAR"] {
			t.Error("expected false for unset env var")
		}
	})

	t.Run("env_var_empty_string", func(t *testing.T) {
		envVars := map[string]string{"EMPTY": ""}
		env := testEnv("linux", "amd64", envVars, nil)
		result, err := ResolvePredicatesWithEnv([]string{"env.EMPTY"}, "/tmp", env)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["env.EMPTY"] {
			t.Error("expected false for empty env var")
		}
	})

	t.Run("env_empty_name_error", func(t *testing.T) {
		env := testEnv("linux", "amd64", nil, nil)
		_, err := ResolvePredicatesWithEnv([]string{"env."}, "/tmp", env)
		if err == nil {
			t.Fatal("expected error for empty env var name")
		}
	})
}

func TestCommandPredicates(t *testing.T) {
	t.Run("command_found", func(t *testing.T) {
		commands := map[string]bool{"git": true}
		env := testEnv("linux", "amd64", nil, commands)
		result, err := ResolvePredicatesWithEnv([]string{"command.git"}, "/tmp", env)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result["command.git"] {
			t.Error("expected true for found command")
		}
	})

	t.Run("command_not_found", func(t *testing.T) {
		env := testEnv("linux", "amd64", nil, nil)
		result, err := ResolvePredicatesWithEnv([]string{"command.nonexistent_tool_xyz"}, "/tmp", env)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["command.nonexistent_tool_xyz"] {
			t.Error("expected false for missing command")
		}
	})

	t.Run("command_empty_name_error", func(t *testing.T) {
		env := testEnv("linux", "amd64", nil, nil)
		_, err := ResolvePredicatesWithEnv([]string{"command."}, "/tmp", env)
		if err == nil {
			t.Fatal("expected error for empty command name")
		}
	})
}

func TestFileExistsPredicate(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "exists.txt"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}

	t.Run("file_exists", func(t *testing.T) {
		env := testEnv("linux", "amd64", nil, nil)
		result, err := ResolvePredicatesWithEnv(
			[]string{`file.exists("exists.txt")`}, dir, env,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result[`file.exists("exists.txt")`] {
			t.Error("expected true for existing file")
		}
	})

	t.Run("file_not_exists", func(t *testing.T) {
		env := testEnv("linux", "amd64", nil, nil)
		result, err := ResolvePredicatesWithEnv(
			[]string{`file.exists("nope.txt")`}, dir, env,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result[`file.exists("nope.txt")`] {
			t.Error("expected false for nonexistent file")
		}
	})

	t.Run("file_exists_is_directory", func(t *testing.T) {
		env := testEnv("linux", "amd64", nil, nil)
		result, err := ResolvePredicatesWithEnv(
			[]string{`file.exists("subdir")`}, dir, env,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result[`file.exists("subdir")`] {
			t.Error("expected false when path is a directory, not a file")
		}
	})
}

func TestDirExistsPredicate(t *testing.T) {
	dir := t.TempDir()

	if err := os.Mkdir(filepath.Join(dir, "mydir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "afile.txt"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("dir_exists", func(t *testing.T) {
		env := testEnv("linux", "amd64", nil, nil)
		result, err := ResolvePredicatesWithEnv(
			[]string{`dir.exists("mydir")`}, dir, env,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result[`dir.exists("mydir")`] {
			t.Error("expected true for existing directory")
		}
	})

	t.Run("dir_not_exists", func(t *testing.T) {
		env := testEnv("linux", "amd64", nil, nil)
		result, err := ResolvePredicatesWithEnv(
			[]string{`dir.exists("nope")`}, dir, env,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result[`dir.exists("nope")`] {
			t.Error("expected false for nonexistent directory")
		}
	})

	t.Run("dir_exists_is_file", func(t *testing.T) {
		env := testEnv("linux", "amd64", nil, nil)
		result, err := ResolvePredicatesWithEnv(
			[]string{`dir.exists("afile.txt")`}, dir, env,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result[`dir.exists("afile.txt")`] {
			t.Error("expected false when path is a file, not a directory")
		}
	})
}

func TestUnknownPredicate(t *testing.T) {
	env := testEnv("linux", "amd64", nil, nil)
	_, err := ResolvePredicatesWithEnv([]string{"bogus.thing"}, "/tmp", env)
	if err == nil {
		t.Fatal("expected error for unknown predicate")
	}
	errStr := err.Error()
	if !containsStr(errStr, "unknown predicate") || !containsStr(errStr, "bogus.thing") {
		t.Errorf("error should mention 'unknown predicate' and the name, got: %s", errStr)
	}
}

func TestMultiplePredicates(t *testing.T) {
	envVars := map[string]string{"SHELL": "/bin/zsh", "CI": "true"}
	commands := map[string]bool{"docker": true}
	env := testEnv("darwin", "arm64", envVars, commands)

	preds := []string{
		"os.macos",
		"os.arch.arm64",
		"shell.zsh",
		"env.CI",
		"command.docker",
	}

	result, err := ResolvePredicatesWithEnv(preds, "/tmp", env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, p := range preds {
		if !result[p] {
			t.Errorf("expected %q to be true", p)
		}
	}
}

func TestEmptyPredicateList(t *testing.T) {
	env := testEnv("linux", "amd64", nil, nil)
	result, err := ResolvePredicatesWithEnv(nil, "/tmp", env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestDefaultRuntimeEnv(t *testing.T) {
	env := DefaultRuntimeEnv()
	if env.GOOS == "" {
		t.Error("GOOS should not be empty")
	}
	if env.GOARCH == "" {
		t.Error("GOARCH should not be empty")
	}
}

func TestResolvePredicatesIntegration(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ResolvePredicates(
		[]string{`file.exists("test.txt")`}, dir,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result[`file.exists("test.txt")`] {
		t.Error("expected true from real ResolvePredicates")
	}
}

func TestValidatePredicateName(t *testing.T) {
	valid := []string{
		"os.macos",
		"os.linux",
		"os.windows",
		"os.arch.arm64",
		"os.arch.amd64",
		"shell.zsh",
		"shell.bash",
		"env.HOME",
		"env.CI",
		"env.MY_CUSTOM_VAR",
		"command.docker",
		"command.brew",
		"command.git",
		`file.exists(".env")`,
		`file.exists("config/settings.yaml")`,
		`dir.exists("src")`,
		`dir.exists("vendor/cache")`,
	}
	for _, name := range valid {
		t.Run("valid_"+name, func(t *testing.T) {
			if err := ValidatePredicateName(name); err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
		})
	}

	invalid := []struct {
		name    string
		wantMsg string
	}{
		{"os.macoss", "unknown predicate"},
		{"os.freebsd", "unknown predicate"},
		{"os.arch.mips", "unknown predicate"},
		{"bogus.thing", "unknown predicate"},
		{"env.", "env variable name is empty"},
		{"command.", "command name is empty"},
		{"shell.fish", "unknown predicate"},
		{"whatisthis", "unknown predicate"},
	}
	for _, tc := range invalid {
		t.Run("invalid_"+tc.name, func(t *testing.T) {
			err := ValidatePredicateName(tc.name)
			if err == nil {
				t.Fatal("expected error")
			}
			if !containsStr(err.Error(), tc.wantMsg) {
				t.Errorf("expected error to contain %q, got: %v", tc.wantMsg, err)
			}
		})
	}
}

func TestValidateConditionExpr(t *testing.T) {
	t.Run("valid_simple", func(t *testing.T) {
		if err := ValidateConditionExpr("os.macos"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("valid_compound", func(t *testing.T) {
		if err := ValidateConditionExpr("os.macos and not command.jq"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("valid_complex", func(t *testing.T) {
		if err := ValidateConditionExpr("(os.linux or os.macos) and command.docker"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("valid_file_exists", func(t *testing.T) {
		if err := ValidateConditionExpr(`file.exists(".env")`); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("valid_empty", func(t *testing.T) {
		if err := ValidateConditionExpr(""); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("syntax_error", func(t *testing.T) {
		err := ValidateConditionExpr("os.macos and and os.linux")
		if err == nil {
			t.Fatal("expected error for syntax error")
		}
	})

	t.Run("unknown_predicate", func(t *testing.T) {
		err := ValidateConditionExpr("os.macoss")
		if err == nil {
			t.Fatal("expected error for unknown predicate")
		}
		if !containsStr(err.Error(), "unknown predicate") {
			t.Errorf("expected error to mention 'unknown predicate', got: %v", err)
		}
	})

	t.Run("unknown_in_compound", func(t *testing.T) {
		err := ValidateConditionExpr("os.macos and bogus.thing")
		if err == nil {
			t.Fatal("expected error for unknown predicate in compound expression")
		}
		if !containsStr(err.Error(), "unknown predicate") {
			t.Errorf("expected error to mention 'unknown predicate', got: %v", err)
		}
	})
}

// --- Evaluator tests ---

func TestEvaluatorShouldSkip(t *testing.T) {
	t.Run("true_condition_does_not_skip", func(t *testing.T) {
		env := testEnv("darwin", "arm64", nil, nil)
		ev := NewEvaluator("/tmp", env)
		skip, err := ev.ShouldSkip("os.macos")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if skip {
			t.Error("expected not skipped for true condition")
		}
	})

	t.Run("false_condition_skips", func(t *testing.T) {
		env := testEnv("linux", "amd64", nil, nil)
		ev := NewEvaluator("/tmp", env)
		skip, err := ev.ShouldSkip("os.macos")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !skip {
			t.Error("expected skipped for false condition")
		}
	})

	t.Run("empty_expression_does_not_skip", func(t *testing.T) {
		env := testEnv("linux", "amd64", nil, nil)
		ev := NewEvaluator("/tmp", env)
		skip, err := ev.ShouldSkip("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if skip {
			t.Error("expected not skipped for empty expression")
		}
	})

	t.Run("complex_condition", func(t *testing.T) {
		commands := map[string]bool{"docker": true}
		env := testEnv("darwin", "arm64", nil, commands)
		ev := NewEvaluator("/tmp", env)
		skip, err := ev.ShouldSkip("os.macos and command.docker")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if skip {
			t.Error("expected not skipped for 'os.macos and command.docker' on darwin with docker")
		}
	})

	t.Run("nil_env_uses_default", func(t *testing.T) {
		ev := NewEvaluator("/tmp", nil)
		_, err := ev.ShouldSkip("os.macos or os.linux")
		if err != nil {
			t.Fatalf("unexpected error with nil env: %v", err)
		}
	})

	t.Run("parse_error", func(t *testing.T) {
		env := testEnv("linux", "amd64", nil, nil)
		ev := NewEvaluator("/tmp", env)
		_, err := ev.ShouldSkip("os.macos and and os.linux")
		if err == nil {
			t.Fatal("expected parse error")
		}
	})

	t.Run("unknown_predicate_error", func(t *testing.T) {
		env := testEnv("linux", "amd64", nil, nil)
		ev := NewEvaluator("/tmp", env)
		_, err := ev.ShouldSkip("bogus.thing")
		if err == nil {
			t.Fatal("expected error for unknown predicate")
		}
	})
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
