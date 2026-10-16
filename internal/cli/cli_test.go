package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sdd.dev/sdd/internal/cli"
	"sdd.dev/sdd/internal/engine"
)

func testEnv(workdir string) cli.Env {
	return cli.Env{
		StdinTerminal:  false,
		StdoutTerminal: false,
		Workdir:        workdir,
		Build:          engine.BuildMetadata{Version: "0.1.0-test", Commit: "test", BuildTime: "unknown", GoVersion: "go1.24.0"},
	}
}

func runCLI(t *testing.T, argv []string, env cli.Env) (stdout, stderr string, exit int) {
	t.Helper()
	var out, errOut bytes.Buffer
	exit = cli.Execute(context.Background(), strings.NewReader(""), &out, &errOut, argv, env)
	return out.String(), errOut.String(), exit
}

func decodeEnvelope(t *testing.T, payload string) map[string]any {
	t.Helper()
	var envelope map[string]any
	if err := json.Unmarshal([]byte(payload), &envelope); err != nil {
		t.Fatalf("CLI JSON is not one envelope: %v\n%s", err, payload)
	}
	return envelope
}

func writeWorkspace(t *testing.T, root string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, ".sdd", "work", "apple"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".sdd", "workspace.yaml"),
		[]byte("formatVersion: 1\nworkspaceId: ws-1\nrevision: 11\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".sdd", "work", "apple", "work.yaml"),
		[]byte("schemaVersion: 1\nworkId: work-1\nslug: apple\nrevision: 5\nstatus: active\ncomposition: quick-spec\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCLIVersionIgnoresWorkspace(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	stdout, stderr, exit := runCLI(t, []string{"version", "--json"}, testEnv(root))
	if exit != 0 {
		t.Fatalf("version exit = %d stderr = %q", exit, stderr)
	}
	envelope := decodeEnvelope(t, stdout)
	if envelope["command"] != "version" || envelope["ok"] != true {
		t.Fatalf("version envelope: %v", envelope)
	}
	result := envelope["result"].(map[string]any)
	if result["version"] != "0.1.0-test" || result["protocolSchemaVersion"] != "1.0" {
		t.Fatalf("version result: %v", result)
	}
	if strings.Count(stdout, "\n") != 1 {
		t.Fatalf("version JSON is not one line: %q", stdout)
	}

	human, humanErr, humanExit := runCLI(t, []string{"version"}, testEnv(root))
	if humanExit != 0 || !strings.HasPrefix(human, "sdd 0.1.0-test") {
		t.Fatalf("version human: exit=%d stdout=%q stderr=%q", humanExit, human, humanErr)
	}
	if humanErr != "" {
		t.Fatalf("successful version wrote stderr: %q", humanErr)
	}

	dash, _, dashExit := runCLI(t, []string{"--version"}, testEnv(root))
	if dashExit != 0 || !strings.HasPrefix(dash, "sdd 0.1.0-test") {
		t.Fatalf("--version: exit=%d stdout=%q", dashExit, dash)
	}
}

func TestCLIHelpNeedsNoWorkspace(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	for _, argv := range [][]string{{"--help"}, {"-h"}, {"help"}, {"help", "status"}, {"status", "--help"}} {
		stdout, stderr, exit := runCLI(t, argv, testEnv(root))
		if exit != 0 {
			t.Fatalf("help %v exit = %d stderr = %q", argv, exit, stderr)
		}
		if !strings.Contains(stdout, "Usage:") {
			t.Fatalf("help %v stdout = %q", argv, stdout)
		}
		if strings.Contains(stdout, "\x1b[") {
			t.Fatalf("help %v contains ANSI", argv)
		}
	}

	_, stderr, exit := runCLI(t, []string{"help", "bogus"}, testEnv(root))
	if exit != 2 || stderr == "" {
		t.Fatalf("unknown help topic: exit=%d stderr=%q", exit, stderr)
	}

	bare, _, bareExit := runCLI(t, nil, testEnv(root))
	if bareExit != 0 || !strings.Contains(bare, "Usage:") {
		t.Fatalf("bare invocation: exit=%d stdout=%q", bareExit, bare)
	}
}

func TestCLIUnknownAndMalformedInput(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	stdout, _, exit := runCLI(t, []string{"bogus", "--json"}, testEnv(root))
	if exit != 2 {
		t.Fatalf("unknown command exit = %d", exit)
	}
	envelope := decodeEnvelope(t, stdout)
	if envelope["command"] != "unknown" || envelope["requestedCommand"] != "bogus" {
		t.Fatalf("unknown envelope: %v", envelope)
	}

	// Leading options must not hide the unknown name.
	flagFirst, _, flagFirstExit := runCLI(t, []string{"--json", "bogus"}, testEnv(root))
	if flagFirstExit != 2 {
		t.Fatalf("flag-first unknown exit = %d", flagFirstExit)
	}
	flagFirstEnvelope := decodeEnvelope(t, flagFirst)
	if flagFirstEnvelope["requestedCommand"] != "bogus" {
		t.Fatalf("flag-first unknown envelope: %v", flagFirstEnvelope)
	}

	_, stderr, usageExit := runCLI(t, []string{"status", "--badflag"}, testEnv(root))
	if usageExit != 2 || stderr == "" {
		t.Fatalf("bad flag: exit=%d stderr=%q", usageExit, stderr)
	}

	_, flagErr, flagExit := runCLI(t, []string{"--badflag"}, testEnv(root))
	if flagExit != 2 || flagErr == "" {
		t.Fatalf("global bad flag: exit=%d stderr=%q", flagExit, flagErr)
	}

	_, conflictErr, conflictExit := runCLI(t, []string{"--json", "--output=text", "version"}, testEnv(root))
	if conflictExit != 2 {
		t.Fatalf("conflicting output: exit=%d stderr=%q", conflictExit, conflictErr)
	}
	_ = conflictErr

	_, modeErr, modeExit := runCLI(t, []string{"--output=bogus", "version"}, testEnv(root))
	if modeExit != 2 || modeErr == "" {
		t.Fatalf("bad output mode: exit=%d stderr=%q", modeExit, modeErr)
	}

	_, tuiErr, tuiExit := runCLI(t, []string{"--output=tui", "version"}, testEnv(root))
	if tuiExit != 2 || tuiErr == "" {
		t.Fatalf("tui mode without TTY: exit=%d stderr=%q", tuiExit, tuiErr)
	}
	ttyEnv := testEnv(root)
	ttyEnv.StdinTerminal = true
	ttyEnv.StdoutTerminal = true
	_, ttyTuiErr, ttyTuiExit := runCLI(t, []string{"--output=tui", "version"}, ttyEnv)
	if ttyTuiExit != 2 || !strings.Contains(ttyTuiErr, "not implemented") {
		t.Fatalf("tui mode with TTY: exit=%d stderr=%q", ttyTuiExit, ttyTuiErr)
	}
}

func TestCLIStatusAndDoctorStreams(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeWorkspace(t, root)

	stdout, stderr, exit := runCLI(t, []string{"status", "--json"}, testEnv(root))
	if exit != 0 {
		t.Fatalf("status exit = %d stderr = %q", exit, stderr)
	}
	envelope := decodeEnvelope(t, stdout)
	if envelope["command"] != "status" {
		t.Fatalf("status envelope: %v", envelope)
	}
	if stderr != "" {
		t.Fatalf("successful status wrote stderr: %q", stderr)
	}

	human, humanErr, humanExit := runCLI(t, []string{"status"}, testEnv(root))
	if humanExit != 0 || !strings.Contains(human, "1 active work item") {
		t.Fatalf("status human: exit=%d stdout=%q stderr=%q", humanExit, human, humanErr)
	}

	_, missingErr, missingExit := runCLI(t, []string{"status"}, testEnv(t.TempDir()))
	if missingExit != 3 || !strings.Contains(missingErr, "DIAG_WORKSPACE_NOT_FOUND") {
		t.Fatalf("missing workspace status: exit=%d stderr=%q", missingExit, missingErr)
	}

	doctorOut, doctorErr, doctorExit := runCLI(t, []string{"doctor", "--json"}, testEnv(root))
	if doctorExit != 0 {
		t.Fatalf("doctor exit = %d stderr = %q", doctorExit, doctorErr)
	}
	doctorEnvelope := decodeEnvelope(t, doctorOut)
	if doctorEnvelope["command"] != "doctor" {
		t.Fatalf("doctor envelope: %v", doctorEnvelope)
	}
}

func TestCLIStartRequiresDryRun(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeWorkspace(t, root)

	_, noFlagErr, noFlagExit := runCLI(t, []string{"start", "add export"}, testEnv(root))
	if noFlagExit != 2 || !strings.Contains(noFlagErr, "--dry-run") {
		t.Fatalf("bare start: exit=%d stderr=%q", noFlagExit, noFlagErr)
	}

	stdout, _, dryExit := runCLI(t, []string{"start", "--dry-run", "add export", "--json"}, testEnv(root))
	if dryExit != 0 {
		t.Fatalf("dry-run exit = %d", dryExit)
	}
	envelope := decodeEnvelope(t, stdout)
	result := envelope["result"].(map[string]any)
	if result["kind"] != "start-dry-run" {
		t.Fatalf("dry-run result: %v", result)
	}
	if plan := result["plan"].(map[string]any); plan["authorizable"] != false {
		t.Fatalf("dry-run plan authorizable: %v", plan)
	}

	_, arityErr, arityExit := runCLI(t, []string{"start", "--dry-run"}, testEnv(root))
	if arityExit != 2 || arityErr == "" {
		t.Fatalf("start without intent: exit=%d stderr=%q", arityExit, arityErr)
	}

	// Tokens after -- are positional, even option-shaped ones.
	terminated, _, terminatedExit := runCLI(t, []string{"start", "--dry-run", "--json", "--", "--version"}, testEnv(root))
	if terminatedExit != 0 {
		t.Fatalf("terminated intent: exit=%d stdout=%q", terminatedExit, terminated)
	}
	terminatedEnvelope := decodeEnvelope(t, terminated)
	plan := terminatedEnvelope["result"].(map[string]any)["plan"].(map[string]any)
	if plan["inputDigest"] != engine.BlobDigest("sdd.start-intent/1", []byte("--version")) {
		t.Fatalf("terminated intent not preserved: %v", plan)
	}

	_, blankErr, blankExit := runCLI(t, []string{"start", "--dry-run", "  "}, testEnv(root))
	if blankExit != 2 || !strings.Contains(blankErr, "intent is required") {
		t.Fatalf("start with blank intent: exit=%d stderr=%q", blankExit, blankErr)
	}
}

func TestCLIRedactsSecretsInBothModes(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	secret := "hunter2token"
	// Unknown-command text is reflected verbatim: it is the caller's own
	// input, and machines need the exact request for attribution. Secret
	// disclosure protection applies to observed content and credential
	// patterns, which are redacted in both renderers.
	jsonOut, _, jsonExit := runCLI(t, []string{"bogus-" + secret, "--json"}, testEnv(root))
	if jsonExit != 2 {
		t.Fatalf("unknown exit = %d", jsonExit)
	}
	if !strings.Contains(jsonOut, "bogus-"+secret) {
		t.Fatalf("unknown request not preserved: %s", jsonOut)
	}
	_, humanErr, _ := runCLI(t, []string{"status", "--root", "/repo/token=" + secret}, testEnv(root))
	if strings.Contains(humanErr, secret) {
		t.Fatalf("CLI diagnostic leaked the credential: %q", humanErr)
	}
	if !strings.Contains(humanErr, "[REDACTED]") {
		t.Fatal("CLI diagnostic lacks redaction markers")
	}
}

func TestCLINoColorMeansNoANSI(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeWorkspace(t, root)
	for _, argv := range [][]string{{"version"}, {"status"}, {"doctor"}} {
		stdout, stderr, exit := runCLI(t, argv, testEnv(root))
		if exit != 0 {
			t.Fatalf("%v exit = %d", argv, exit)
		}
		if strings.Contains(stdout+stderr, "\x1b[") {
			t.Fatalf("%v emitted ANSI escapes", argv)
		}
	}
}

func TestCLIWithoutWorkdirKeepsWorkspaceFreeCommands(t *testing.T) {
	t.Parallel()

	env := testEnv("")
	stdout, _, exit := runCLI(t, []string{"version"}, env)
	if exit != 0 || !strings.HasPrefix(stdout, "sdd ") {
		t.Fatalf("workdir-less version: exit=%d stdout=%q", exit, stdout)
	}
	helpOut, _, helpExit := runCLI(t, []string{"--help"}, env)
	if helpExit != 0 || !strings.Contains(helpOut, "Usage:") {
		t.Fatalf("workdir-less help: exit=%d stdout=%q", helpExit, helpOut)
	}
	_, _, statusExit := runCLI(t, []string{"status"}, env)
	if statusExit == 0 {
		t.Fatal("workdir-less status unexpectedly succeeded")
	}
}

func TestCLICancelledContextInterrupts(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var stdout, stderr bytes.Buffer
	exit := cli.Execute(ctx, strings.NewReader(""), &stdout, &stderr, []string{"status"}, testEnv(root))
	if exit != 130 {
		t.Fatalf("cancelled exit = %d stdout=%q stderr=%q", exit, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "DIAG_INTERRUPTED") {
		t.Fatalf("cancelled stderr = %q", stderr.String())
	}
}
