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
	p.InstallStatus(StatusSuccessCached, "my-tool", "already installed", "1.2.3")
	got := buf.String()
	if !strings.Contains(got, "✓") || !strings.Contains(got, "already installed") || !strings.Contains(got, "1.2.3") {
		t.Errorf("unexpected output: %q", got)
	}
}

func TestInstallStatus_AlreadyInstalled_WithColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, true)
	p.InstallStatus(StatusSuccessCached, "my-tool", "already installed", "1.2.3")
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
	p.InstallStatus(StatusSuccess, "my-tool", "installed", "2.0.0")
	got := buf.String()
	if !strings.Contains(got, boldGreen) {
		t.Errorf("newly installed should use bold green, got %q", got)
	}
}

func TestInstallStatus_Installing_NoStyle(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, true)
	p.InstallStatus(StatusInProgress, "my-tool", "installing...", "")
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
	p.InstallStatus(StatusFailed, "my-tool", "failed", "")
	got := buf.String()
	if !strings.Contains(got, boldRed) {
		t.Errorf("failed should use bold red, got %q", got)
	}
}

func TestInstallStatus_NoVersion(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, false)
	p.InstallStatus(StatusSuccess, "my-tool", "installed", "")
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

func TestNewForWriter_BufferDisablesColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewForWriter(&buf)
	p.Green("test")
	got := buf.String()
	if strings.Contains(got, "\033[") {
		t.Errorf("buffer writer should disable color, got %q", got)
	}
	if got != "test" {
		t.Errorf("expected plain text, got %q", got)
	}
}

func TestNewForWriter_OsFileUsesTTYDetection(t *testing.T) {
	f, err := os.CreateTemp("", "display-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	p := NewForWriter(f)
	p.Green("test")

	if _, seekErr := f.Seek(0, 0); seekErr != nil {
		t.Fatal(seekErr)
	}
	content := make([]byte, 256)
	n, _ := f.Read(content)
	got := string(content[:n])
	if strings.Contains(got, "\033[") {
		t.Errorf("non-TTY file should disable color, got %q", got)
	}
}

func TestNewForWriter_WritesCorrectOutput(t *testing.T) {
	var buf bytes.Buffer
	p := NewForWriter(&buf)
	p.Dim("dimmed")
	p.Bold("bolded")
	p.Red("error")
	got := buf.String()
	if got != "dimmedbolded" + "error" {
		t.Errorf("expected concatenated plain text, got %q", got)
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

func TestPackageFetch_Cached(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, false)
	p.PackageFetch(StatusSuccessCached, "yotam180/pi-common@v1.2", "cached", "")
	got := buf.String()
	if !strings.Contains(got, "✓") {
		t.Errorf("expected ✓ icon, got %q", got)
	}
	if !strings.Contains(got, "cached") {
		t.Errorf("expected 'cached' status, got %q", got)
	}
	if !strings.Contains(got, "yotam180/pi-common@v1.2") {
		t.Errorf("expected source name, got %q", got)
	}
}

func TestPackageFetch_WithDetail(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, false)
	p.PackageFetch(StatusSuccessCached, "file:~/my-automations", "found", "alias: mytools")
	got := buf.String()
	if !strings.Contains(got, "alias: mytools") {
		t.Errorf("expected detail in output, got %q", got)
	}
}

func TestPackageFetch_Fetching(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, false)
	p.PackageFetch(StatusInProgress, "testorg/pkg@v1.0", "fetching...", "")
	got := buf.String()
	if !strings.Contains(got, "↓") {
		t.Errorf("expected ↓ icon, got %q", got)
	}
	if !strings.Contains(got, "fetching...") {
		t.Errorf("expected 'fetching...' status, got %q", got)
	}
}

func TestPackageFetch_Failed(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, false)
	p.PackageFetch(StatusFailed, "badorg/badpkg@v0.1", "failed", "")
	got := buf.String()
	if !strings.Contains(got, "✗") {
		t.Errorf("expected ✗ icon, got %q", got)
	}
	if !strings.Contains(got, "failed") {
		t.Errorf("expected 'failed' status, got %q", got)
	}
}

func TestPackageFetch_Warning_NoColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, false)
	p.PackageFetch(StatusWarning, "file:~/missing-path", "not found", "")
	got := buf.String()
	if !strings.Contains(got, "⚠") {
		t.Errorf("expected ⚠ icon, got %q", got)
	}
	if !strings.Contains(got, "not found") {
		t.Errorf("expected 'not found' status, got %q", got)
	}
}

func TestPackageFetch_Warning_WithColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, true)
	p.PackageFetch(StatusWarning, "file:~/missing-path", "not found", "")
	got := buf.String()
	if !strings.Contains(got, yellow) {
		t.Errorf("warning icon should use yellow style, got %q", got)
	}
	if strings.Contains(got, boldRed) || strings.Contains(got, dim) {
		t.Errorf("warning icon should not use red or dim style, got %q", got)
	}
	if !strings.Contains(got, "not found") {
		t.Errorf("expected 'not found' status, got %q", got)
	}
}

func TestStatusIcon_AllKinds(t *testing.T) {
	tests := []struct {
		kind StatusKind
		icon string
	}{
		{StatusSuccess, "✓"},
		{StatusSuccessCached, "✓"},
		{StatusInProgress, "→"},
		{StatusFailed, "✗"},
		{StatusWarning, "⚠"},
	}
	for _, tt := range tests {
		t.Run(string(tt.kind), func(t *testing.T) {
			got := statusIcon(tt.kind)
			if got != tt.icon {
				t.Errorf("statusIcon(%q) = %q, want %q", tt.kind, got, tt.icon)
			}
		})
	}
}

func TestStatusIcon_Unknown(t *testing.T) {
	got := statusIcon("unknown_kind")
	if got != "?" {
		t.Errorf("statusIcon(unknown) = %q, want %q", got, "?")
	}
}

func TestInstallStatus_Warning_WithColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, true)
	p.InstallStatus(StatusWarning, "my-tool", "partially installed", "")
	got := buf.String()
	if !strings.Contains(got, yellow) {
		t.Errorf("warning should use yellow style, got %q", got)
	}
	if !strings.Contains(got, "⚠") {
		t.Errorf("expected ⚠ icon, got %q", got)
	}
}

func TestPackageFetch_InProgress_UsesDownArrow(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, false)
	p.PackageFetch(StatusInProgress, "org/repo@v1.0", "fetching...", "")
	got := buf.String()
	if !strings.Contains(got, "↓") {
		t.Errorf("PackageFetch with StatusInProgress should use ↓ icon, got %q", got)
	}
}

func TestPrintStatusLine_AllKinds(t *testing.T) {
	tests := []struct {
		kind  StatusKind
		color string
	}{
		{StatusFailed, boldRed},
		{StatusSuccessCached, dim},
		{StatusSuccess, boldGreen},
		{StatusWarning, yellow},
	}
	for _, tt := range tests {
		t.Run(string(tt.kind), func(t *testing.T) {
			var buf bytes.Buffer
			p := NewWithColor(&buf, true)
			p.printStatusLine(tt.kind, "test line\n")
			got := buf.String()
			if !strings.Contains(got, tt.color) {
				t.Errorf("printStatusLine(%q) should contain %q ANSI code, got %q", tt.kind, tt.color, got)
			}
		})
	}
}

func TestPrintStatusLine_InProgress_Plain(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, true)
	p.printStatusLine(StatusInProgress, "test line\n")
	got := buf.String()
	if strings.Contains(got, "\033[") {
		t.Errorf("StatusInProgress should use plain style (no ANSI codes), got %q", got)
	}
	if got != "test line\n" {
		t.Errorf("expected plain text, got %q", got)
	}
}

func TestWarn_NoColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, false)
	p.Warn("something went wrong: %s\n", "details")
	got := buf.String()
	if got != "something went wrong: details\n" {
		t.Errorf("expected plain text, got %q", got)
	}
}

func TestWarn_WithColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewWithColor(&buf, true)
	p.Warn("something went wrong\n")
	got := buf.String()
	if !strings.Contains(got, yellow) {
		t.Errorf("expected yellow ANSI code, got %q", got)
	}
	if !strings.Contains(got, "something went wrong") {
		t.Errorf("expected warning text, got %q", got)
	}
	if !strings.HasSuffix(got, reset) {
		t.Errorf("expected reset suffix, got %q", got)
	}
}
