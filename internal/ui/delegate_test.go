package ui

import (
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
)

func TestTruncateServiceNamePreservesSuffix(t *testing.T) {
	name := "very-long-worker-name-that-needs-truncation.service"

	got := truncateServiceName(name, 24)

	if !strings.HasSuffix(got, ".service") {
		t.Fatalf("expected suffix to be preserved, got %q", got)
	}
	if width := runewidth.StringWidth(got); width > 24 {
		t.Fatalf("expected width <= 24, got %d for %q", width, got)
	}
}

func TestTruncateServiceNameHandlesUnicode(t *testing.T) {
	name := "日本語-worker-with-a-long-name.service"

	got := truncateServiceName(name, 18)

	if width := runewidth.StringWidth(got); width > 18 {
		t.Fatalf("expected width <= 18, got %d for %q", width, got)
	}
	if strings.ContainsRune(got, '\uFFFD') {
		t.Fatalf("expected valid unicode without replacement characters, got %q", got)
	}
}

func TestTruncateDescriptionHandlesUnicode(t *testing.T) {
	desc := "syncs 日本語 files and emits detailed logs"

	got := truncateDescription(desc, 16)

	if width := runewidth.StringWidth(got); width > 16 {
		t.Fatalf("expected width <= 16, got %d for %q", width, got)
	}
	if strings.ContainsRune(got, '\uFFFD') {
		t.Fatalf("expected valid unicode without replacement characters, got %q", got)
	}
}
