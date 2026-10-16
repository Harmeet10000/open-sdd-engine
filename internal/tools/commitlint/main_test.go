package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestValidSubject(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		subject string
		want    bool
	}{
		{name: "feature", subject: "feat: add output negotiation", want: true},
		{name: "scoped fix", subject: "fix(cli): preserve JSON stdout", want: true},
		{name: "breaking", subject: "feat(protocol)!: publish version 2", want: true},
		{name: "dependabot", subject: "chore(deps): bump cobra to v1.10.2", want: true},
		{name: "missing type", subject: "add output negotiation", want: false},
		{name: "uppercase type", subject: "Feat: add output negotiation", want: false},
		{name: "empty summary", subject: "fix: ", want: false},
		{name: "leading summary space", subject: "fix:  broken", want: false},
		{name: "trailing period", subject: "fix: reject ambiguous input.", want: false},
		{name: "too long", subject: "feat: " + strings.Repeat("x", 95), want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := validSubject(test.subject); got != test.want {
				t.Fatalf("validSubject(%q) = %t, want %t", test.subject, got, test.want)
			}
		})
	}
}

func TestRunReportsEveryInvalidSubject(t *testing.T) {
	t.Parallel()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := run(context.Background(), &stdout, &stderr, options{message: "not conventional"})
	if err == nil {
		t.Fatal("invalid subject was accepted")
	}
	if !strings.Contains(stderr.String(), `invalid commit subject: "not conventional"`) {
		t.Fatalf("stderr did not identify the invalid subject: %q", stderr.String())
	}
}

func TestRunAcceptsOneMessage(t *testing.T) {
	t.Parallel()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := run(context.Background(), &stdout, &stderr, options{message: "ci: add contribution gates"})
	if err != nil {
		t.Fatal(err)
	}
	if got := stdout.String(); got != "validated 1 commit subject(s)\n" {
		t.Fatalf("stdout = %q", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("unexpected stderr: %q", stderr.String())
	}
}

func TestLoadSubjectsRejectsAmbiguousInput(t *testing.T) {
	t.Parallel()

	_, err := loadSubjects(context.Background(), options{
		message: "fix: one subject",
		base:    "main",
	})
	if err == nil || !strings.Contains(err.Error(), "either -message or -base") {
		t.Fatalf("loadSubjects() error = %v", err)
	}
}

func TestLoadSubjectsRequiresInput(t *testing.T) {
	t.Parallel()

	_, err := loadSubjects(context.Background(), options{})
	if err == nil || !strings.Contains(err.Error(), "is required") {
		t.Fatalf("loadSubjects() error = %v", err)
	}
}

func TestLoadSubjectsReturnsOneMessage(t *testing.T) {
	t.Parallel()

	subjects, err := loadSubjects(context.Background(), options{message: "docs: explain gates"})
	if err != nil {
		t.Fatal(err)
	}
	if len(subjects) != 1 || subjects[0] != "docs: explain gates" {
		t.Fatalf("loadSubjects() = %#v", subjects)
	}
}

func TestLoadSubjectsRejectsUnsafeRevision(t *testing.T) {
	t.Parallel()

	_, err := loadSubjects(context.Background(), options{base: "--all", head: "HEAD"})
	if err == nil || !strings.Contains(err.Error(), "simple Git revision") {
		t.Fatalf("loadSubjects() error = %v", err)
	}
}

func TestValidRevision(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		revision string
		want     bool
	}{
		{name: "SHA", revision: "0123456789abcdef", want: true},
		{name: "remote branch", revision: "origin/main", want: true},
		{name: "head", revision: "HEAD", want: true},
		{name: "option", revision: "--all", want: false},
		{name: "range", revision: "main..feature", want: false},
		{name: "expression", revision: "HEAD~1", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := validRevision(test.revision); got != test.want {
				t.Fatalf("validRevision(%q) = %t, want %t", test.revision, got, test.want)
			}
		})
	}
}
