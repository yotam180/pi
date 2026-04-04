package executor

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
	if got := err.Error(); !contains(got, "unknown predicate") || !contains(got, "bogus.thing") {
		t.Errorf("error should mention 'unknown predicate' and the name, got: %s", got)
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

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
