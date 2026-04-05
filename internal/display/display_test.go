package display

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestPlain_WritesUnformattedText(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, false)
	p.Plain("hello %s", "world")
	if got := buf.String(); got != "hello world" {
		t.Errorf("expected %q, got %q", "hello world", got)
	}
}

func TestPlain_WithColorEnabled(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, true)
	p.Plain("no style %d", 42)
	if got := buf.String(); got != "no style 42" {
		t.Errorf("Plain should never add ANSI codes, got %q", got)
	}
}

func TestDim_NoColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, false)
	p.Dim("faded text")
	if got := buf.String(); got != "faded text" {
		t.Errorf("expected plain text, got %q", got)
	}
}

func TestDim_WithColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, true)
	p.Dim("faded text")
	got := buf.String()
	if !strings.Contains(got, "\033[2m") {
		t.Errorf("expected dim ANSI code, got %q", got)
	}
	if !strings.Contains(got, "faded text") {
		t.Errorf("expected text content, got %q", got)
	}
	if !strings.HasSuffix(got, reset) {
		t.Errorf("expected reset suffix, got %q", got)
	}
}

func TestGreen_NoColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, false)
	p.Green("success")
	if got := buf.String(); got != "success" {
		t.Errorf("expected plain text, got %q", got)
	}
}

func TestGreen_WithColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, true)
	p.Green("success")
	got := buf.String()
	if !strings.Contains(got, boldGreen) {
		t.Errorf("expected bold green ANSI code, got %q", got)
	}
	if !strings.Contains(got, "success") {
		t.Errorf("expected text content, got %q", got)
	}
}

func TestRed_NoColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, false)
	p.Red("failure")
	if got := buf.String(); got != "failure" {
		t.Errorf("expected plain text, got %q", got)
	}
}

func TestRed_WithColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, true)
	p.Red("failure")
	got := buf.String()
	if !strings.Contains(got, boldRed) {
		t.Errorf("expected bold red ANSI code, got %q", got)
	}
}

func TestBold_NoColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, false)
	p.Bold("important")
	if got := buf.String(); got != "important" {
		t.Errorf("expected plain text, got %q", got)
	}
}

func TestBold_WithColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, true)
	p.Bold("important")
	got := buf.String()
	if !strings.Contains(got, bold) {
		t.Errorf("expected bold ANSI code, got %q", got)
	}
}

func TestInstallStatus_AlreadyInstalled_NoColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, false)
	p.InstallStatus("✓", "my-tool", "already installed", "1.2.3")
	got := buf.String()
	if !strings.Contains(got, "✓") || !strings.Contains(got, "already installed") || !strings.Contains(got, "1.2.3") {
		t.Errorf("unexpected output: %q", got)
	}
}

func TestInstallStatus_AlreadyInstalled_WithColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, true)
	p.InstallStatus("✓", "my-tool", "already installed", "1.2.3")
	got := buf.String()
	if !strings.Contains(got, dim) {
		t.Errorf("already-installed should use dim style, got %q", got)
	}
	if strings.Contains(got, boldGreen) {
		t.Errorf("already-installed should NOT use bold green, got %q", got)
	}
}

func TestInstallStatus_Installed_WithColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, true)
	p.InstallStatus("✓", "my-tool", "installed", "2.0.0")
	got := buf.String()
	if !strings.Contains(got, boldGreen) {
		t.Errorf("newly installed should use bold green, got %q", got)
	}
}

func TestInstallStatus_Installing_NoStyle(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, true)
	p.InstallStatus("→", "my-tool", "installing...", "")
	got := buf.String()
	if strings.Contains(got, dim) || strings.Contains(got, boldGreen) || strings.Contains(got, boldRed) {
		t.Errorf("installing should use plain style, got %q", got)
	}
	if !strings.Contains(got, "installing...") {
		t.Errorf("expected installing text, got %q", got)
	}
}

func TestInstallStatus_Failed_WithColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, true)
	p.InstallStatus("✗", "my-tool", "failed", "")
	got := buf.String()
	if !strings.Contains(got, boldRed) {
		t.Errorf("failed should use bold red, got %q", got)
	}
}

func TestInstallStatus_NoVersion(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, false)
	p.InstallStatus("✓", "my-tool", "installed", "")
	got := buf.String()
	if strings.Contains(got, "()") {
		t.Errorf("empty version should not produce parens, got %q", got)
	}
	if !strings.Contains(got, "installed") {
		t.Errorf("expected installed text, got %q", got)
	}
}

func TestSetupHeader_Dim(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, true)
	p.SetupHeader("==> setup[0]: docker/up\n")
	got := buf.String()
	if !strings.Contains(got, dim) {
		t.Errorf("setup header should use dim style, got %q", got)
	}
	if !strings.Contains(got, "docker/up") {
		t.Errorf("expected header text, got %q", got)
	}
}

func TestSetupHeader_NoColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, false)
	p.SetupHeader("==> setup[0]: docker/up\n")
	got := buf.String()
	if got != "==> setup[0]: docker/up\n" {
		t.Errorf("expected plain text, got %q", got)
	}
}

func TestNew_BufferWriter_NoColor(t *testing.T) {
	var buf bytes.Buffer
	p := New(&buf)
	p.Green("test")
	got := buf.String()
	if strings.Contains(got, "\033[") {
		t.Errorf("buffer writer should disable color, got %q", got)
	}
}

func TestShouldColor_NoColorEnvVar(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	if shouldColor(os.Stderr) {
		t.Error("shouldColor should return false when NO_COLOR is set")
	}
}

func TestShouldColor_NoColorEmpty(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	if shouldColor(os.Stderr) {
		t.Error("shouldColor should return false when NO_COLOR is set to empty (presence matters)")
	}
}

func TestShouldColor_NonFileWriter(t *testing.T) {
	var buf bytes.Buffer
	if shouldColor(&buf) {
		t.Error("shouldColor should return false for non-*os.File writers")
	}
}

func TestFormatArgs(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, false)
	p.Dim("count: %d, name: %s\n", 5, "foo")
	if got := buf.String(); got != "count: 5, name: foo\n" {
		t.Errorf("expected formatted text, got %q", got)
	}
}

func TestStyled_ResetSuffix(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, true)
	p.Red("err")
	got := buf.String()
	if !strings.HasSuffix(got, reset) {
		t.Errorf("styled output should end with reset, got %q", got)
	}
}

func TestMultipleStyles_Sequential(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, true)
	p.Green("ok")
	p.Red("fail")
	p.Dim("skip")
	got := buf.String()
	if strings.Count(got, reset) != 3 {
		t.Errorf("expected 3 resets for 3 styled calls, got %q", got)
	}
}

func TestStepTrace_NoColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, false)
	p.StepTrace("bash", "echo hello")
	got := buf.String()
	if got != "  → bash: echo hello\n" {
		t.Errorf("expected trace line, got %q", got)
	}
}

func TestStepTrace_WithColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, true)
	p.StepTrace("bash", "echo hello")
	got := buf.String()
	if !strings.Contains(got, dim) {
		t.Errorf("trace should use dim style, got %q", got)
	}
	if !strings.Contains(got, "bash: echo hello") {
		t.Errorf("trace should contain command, got %q", got)
	}
}

func TestStepTrace_Truncation(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, false)
	long := strings.Repeat("x", 100)
	p.StepTrace("bash", long)
	got := buf.String()
	if !strings.HasSuffix(strings.TrimSpace(got), "...") {
		t.Errorf("long trace should be truncated, got %q", got)
	}
}

func TestStepTrace_MultilineCollapse(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, false)
	p.StepTrace("bash", "line1\nline2\nline3")
	got := buf.String()
	if strings.Contains(got, "line2") {
		t.Errorf("multiline should be collapsed, got %q", got)
	}
	if !strings.Contains(got, "line1...") {
		t.Errorf("should show first line with ..., got %q", got)
	}
}

func TestTruncateTrace(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"short", "echo hello", 80, "echo hello"},
		{"exact", "abcde", 5, "abcde"},
		{"long", "abcdefghij", 8, "abcde..."},
		{"multiline", "line1\nline2", 80, "line1..."},
		{"multiline long", strings.Repeat("x", 100) + "\nline2", 80, strings.Repeat("x", 77) + "..."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateTrace(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateTrace(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}
