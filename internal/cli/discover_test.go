package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintOnDemandAdvisory_OutputFormat(t *testing.T) {
	var buf bytes.Buffer
	printOnDemandAdvisory(&buf, "yotam180/pi-common@v1.2")

	output := buf.String()

	if !strings.Contains(output, "yotam180/pi-common@v1.2") {
		t.Errorf("expected source in output, got: %q", output)
	}
	if !strings.Contains(output, "tip:") {
		t.Errorf("expected 'tip:' in advisory, got: %q", output)
	}
	if !strings.Contains(output, "packages:") {
		t.Errorf("expected 'packages:' snippet in advisory, got: %q", output)
	}
	if !strings.Contains(output, "- yotam180/pi-common@v1.2") {
		t.Errorf("expected ready-to-paste snippet, got: %q", output)
	}
}

func TestPrintOnDemandAdvisory_NilWriter(t *testing.T) {
	// Should not panic
	printOnDemandAdvisory(nil, "org/repo@v1.0")
}

func TestPrintOnDemandAdvisory_ContainsFetchStatus(t *testing.T) {
	var buf bytes.Buffer
	printOnDemandAdvisory(&buf, "org/repo@v2.0")

	output := buf.String()

	if !strings.Contains(output, "fetched (on demand)") {
		t.Errorf("expected 'fetched (on demand)' status, got: %q", output)
	}
}

func TestPrintOnDemandAdvisory_ContainsDownArrow(t *testing.T) {
	var buf bytes.Buffer
	printOnDemandAdvisory(&buf, "org/repo@v1.0")

	output := buf.String()

	if !strings.Contains(output, "↓") {
		t.Errorf("expected down arrow icon, got: %q", output)
	}
}
