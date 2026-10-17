package help_test

import (
	"strings"
	"testing"

	"sdd.dev/sdd/internal/presentation/help"
)

func TestHelpNamesOnlyImplementedCommands(t *testing.T) {
	t.Parallel()

	text := help.Text
	for _, command := range []string{"version", "doctor", "status", "start"} {
		if !strings.Contains(text, command) {
			t.Fatalf("help text does not name %q", command)
		}
		topic, ok := help.Topic(command)
		if !ok || !strings.Contains(topic, "Usage:") {
			t.Fatalf("command %q lacks a usage topic", command)
		}
	}
	if strings.Contains(text, "\x1b[") {
		t.Fatal("help text contains ANSI escapes")
	}
	if !strings.HasSuffix(text, "\n") || strings.Contains(text, "\t") {
		t.Fatal("help text must end in a newline and use spaces, not tabs")
	}
}

func TestDoctorHelpUsesStableExitTaxonomy(t *testing.T) {
	t.Parallel()

	if !strings.Contains(help.DoctorHelp, "Exits 3") || !strings.Contains(help.DoctorHelp, "exits 6") {
		t.Fatalf("doctor help does not distinguish workspace and policy exits:\n%s", help.DoctorHelp)
	}
}
