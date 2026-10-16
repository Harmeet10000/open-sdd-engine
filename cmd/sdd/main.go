// Command sdd is the native executable for spec-driven development. Main
// only observes the process boundary (terminal facts, workdir, build
// metadata, signals) and delegates to the cli composition root, which
// returns the process exit code. Workflow decisions stay behind
// Engine.Execute; main decides nothing.
package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"

	"sdd.dev/sdd/internal/cli"
	"sdd.dev/sdd/internal/engine"
)

var (
	version   = "0.1.0-dev"
	commit    = "unknown"
	buildTime = "unknown"
)

func main() {
	// A lost working directory must not take down workspace-free commands:
	// pass through empty and let the composition root substitute its
	// documented fallback, which version and help ignore.
	workdir, err := os.Getwd()
	if err != nil {
		workdir = ""
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	code := cli.Execute(ctx, os.Stdin, os.Stdout, os.Stderr, os.Args[1:], cli.Env{
		StdinTerminal:  isTerminal(os.Stdin),
		StdoutTerminal: isTerminal(os.Stdout),
		Workdir:        workdir,
		Build: engine.BuildMetadata{
			Version:   version,
			Commit:    commit,
			BuildTime: buildTime,
			GoVersion: runtime.Version(),
		},
	})
	stop()
	os.Exit(code)
}

func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
