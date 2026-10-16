// Command commitlint validates Conventional Commit subjects without requiring
// a Node.js toolchain in the Go contribution path.
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var conventionalSubject = regexp.MustCompile(
	`^(build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test)(\([a-z0-9][a-z0-9._/-]*\))?!?: [^[:space:]].+$`,
)

var revisionName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]{0,199}$`)

type options struct {
	message string
	base    string
	head    string
}

func main() {
	opts := options{}
	flag.StringVar(&opts.message, "message", "", "validate one commit subject")
	flag.StringVar(&opts.base, "base", "", "validate commits after this revision")
	flag.StringVar(&opts.head, "head", "HEAD", "validate commits through this revision")
	flag.Parse()

	if err := run(context.Background(), os.Stdout, os.Stderr, opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, stdout io.Writer, stderr io.Writer, opts options) error {
	subjects, err := loadSubjects(ctx, opts)
	if err != nil {
		return err
	}

	invalid := make([]string, 0)
	for _, subject := range subjects {
		if !validSubject(subject) {
			invalid = append(invalid, subject)
			fmt.Fprintf(stderr, "invalid commit subject: %q\n", subject)
		}
	}
	if len(invalid) > 0 {
		return fmt.Errorf(
			"%d commit subject(s) violate Conventional Commits; expected type(scope): summary",
			len(invalid),
		)
	}

	fmt.Fprintf(stdout, "validated %d commit subject(s)\n", len(subjects))
	return nil
}

func loadSubjects(ctx context.Context, opts options) ([]string, error) {
	if opts.message != "" {
		if opts.base != "" {
			return nil, errors.New("choose either -message or -base")
		}
		return []string{opts.message}, nil
	}
	if opts.base == "" {
		return nil, errors.New("-message or -base is required")
	}
	if !validRevision(opts.base) || !validRevision(opts.head) {
		return nil, errors.New("base and head must be simple Git revision names")
	}

	revisionRange := opts.base + ".." + opts.head
	// #nosec G204 -- validated revisions are passed as one argv value to a fixed
	// executable; no shell parses or expands their contents.
	command := exec.CommandContext(ctx, "git", "log", "--format=%s", revisionRange)
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("read commit subjects for %q: %w", revisionRange, err)
	}

	subjects := make([]string, 0)
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		subject := strings.TrimSpace(scanner.Text())
		if subject != "" {
			subjects = append(subjects, subject)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan commit subjects: %w", err)
	}
	if len(subjects) == 0 {
		return nil, fmt.Errorf("revision range %q contains no commits", revisionRange)
	}
	return subjects, nil
}

func validRevision(revision string) bool {
	return revisionName.MatchString(revision) && !strings.Contains(revision, "..")
}

func validSubject(subject string) bool {
	if len(subject) > 100 || strings.HasSuffix(subject, ".") {
		return false
	}
	return conventionalSubject.MatchString(subject)
}
