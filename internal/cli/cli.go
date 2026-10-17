// Package cli is the composition root: argv parsing, engine composition,
// stream separation, and exit-code translation. It decides no workflow
// policy and prompts never; every outcome comes from Engine and every
// failure renders through the same human/JSON seam.
package cli

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"sdd.dev/sdd/internal/engine"
	"sdd.dev/sdd/internal/presentation/help"
	"sdd.dev/sdd/internal/presentation/outputmode"
	"sdd.dev/sdd/internal/presentation/render"
)

// Env carries facts the executable boundary observed. Nothing here is read
// from process globals inside this package, so tests can substitute every
// field deterministically.
type Env struct {
	StdinTerminal  bool
	StdoutTerminal bool
	Workdir        string
	Build          engine.BuildMetadata
}

// Execute runs argv (without the program name) and returns the process exit
// code. It never calls os.Exit, never prompts, and never requires a TTY:
// redirected streams receive the same outcomes through the text renderer.
func Execute(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer, argv []string, env Env) int {
	workdir := env.Workdir
	if workdir == "" {
		// The executable boundary could not determine a workdir (getwd
		// failed). Version and help do not observe the filesystem and
		// keep working; workspace commands fail closed downstream.
		workdir = "."
	}
	eng, err := engine.NewEngine(engine.Dependencies{
		Facts:   engine.OSFileFacts{},
		IDs:     engine.RandomIDSource{},
		Policy:  engine.LocalPolicy(),
		Workdir: workdir,
		Build:   env.Build,
	})
	if err != nil {
		// Composition passed incomplete dependencies. Fall back to a
		// literal-built engine so even this failure renders as an outcome.
		fallback, fallbackErr := engine.NewEngine(engine.Dependencies{
			Facts:   engine.OSFileFacts{},
			IDs:     engine.RandomIDSource{},
			Policy:  engine.LocalPolicy(),
			Workdir: ".",
			Build:   engine.DevelopmentBuild(),
		})
		if fallbackErr != nil {
			fmt.Fprintln(stderr, "engine unavailable")
			return 1
		}
		return renderOutcomeFor(stdout, stderr, fallback, jsonFlagOf(argv), outputFlagOf(argv), env,
			fallback.UsageError(engine.Command{}, "invalid engine configuration"))
	}

	if len(argv) > 0 && argv[0] == "help" {
		topic := ""
		if len(argv) > 1 {
			topic = argv[1]
		}
		return printHelp(stdout, stderr, topic)
	}
	for _, token := range preFlagTokens(argv) {
		if token == "--version" || token == "--version=true" {
			return renderOutcomeFor(stdout, stderr, eng, jsonFlagOf(argv), outputFlagOf(argv), env,
				eng.Execute(ctx, engine.NewVersionCommand()))
		}
		if token == "-h" || token == "--help" {
			return printHelp(stdout, stderr, helpTopic(argv))
		}
	}

	exit := 0
	root := newRootCommand(env, eng, &exit, stdin, stdout, stderr)
	root.SetContext(ctx)
	root.SetArgs(argv)
	if err := root.Execute(); err != nil {
		return cliFailure(stdout, stderr, argv, env, eng)
	}
	return exit
}

func printHelp(stdout, stderr io.Writer, topic string) int {
	text := help.Text
	if topic != "" {
		detail, ok := help.Topic(topic)
		if !ok {
			fmt.Fprintf(stderr, "unknown help topic %q\n", topic)
			return 2
		}
		text = detail
	}
	if _, err := io.WriteString(stdout, text); err != nil {
		fmt.Fprintln(stderr, "help output write failed")
		return 1
	}
	return 0
}

func helpTopic(argv []string) string {
	for _, token := range preFlagTokens(argv) {
		if _, ok := help.Topic(token); ok {
			return token
		}
	}
	return ""
}

// preFlagTokens yields argv tokens before a bare "--" terminator. Everything
// after "--" is positional input (an intent may literally be "--version"),
// never an option.
func preFlagTokens(argv []string) []string {
	for index, token := range argv {
		if token == "--" {
			return argv[:index]
		}
	}
	return argv
}

// writeFailure reports a transport failure without echoing the underlying
// error: transport errors can carry paths and descriptors that do not belong
// in machine-readable streams. Consuming err here also documents that the
// failure was handled, not swallowed.
func writeFailure(stderr io.Writer, err error) int {
	if err != nil {
		fmt.Fprintln(stderr, "output write failed")
		return 1
	}
	return 0
}

func newRootCommand(env Env, eng *engine.Engine, exit *int, stdin io.Reader, stdout, stderr io.Writer) *cobra.Command {
	var jsonFlag bool
	var outputFlag string
	var rootFlag string

	run := func(cmd *cobra.Command, build engine.Command) {
		ctx := cmd.Context()
		outcome := eng.Execute(ctx, build)
		*exit = renderOutcomeFor(stdout, stderr, eng, jsonFlag, outputFlag, env, outcome)
	}

	root := &cobra.Command{
		Use:           "sdd",
		Short:         "Deterministic engine for spec-driven development",
		Long:          help.Text,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if _, err := io.WriteString(cmd.OutOrStdout(), help.Text); err != nil {
				*exit = writeFailure(cmd.ErrOrStderr(), err)
				return nil
			}
			*exit = 0
			return nil
		},
	}
	root.CompletionOptions.DisableDefaultCmd = true
	root.SetIn(stdin)
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Emit exactly one protocol envelope on stdout")
	root.PersistentFlags().StringVar(&outputFlag, "output", "", "Output mode: json or text")
	root.PersistentFlags().StringVar(&rootFlag, "root", "", "Inspect only the given workspace root")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print build metadata and the protocol schema version",
		Long:  help.VersionHelp,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			run(cmd, engine.NewVersionCommand())
			return nil
		},
	}

	doctorCmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check executable, workspace, schema, and policy status",
		Long:  help.DoctorHelp,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			run(cmd, engine.NewDoctorCommand(rootFlag))
			return nil
		},
	}

	var workFlag string
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Report zero, one, or a selected work item",
		Long:  help.StatusHelp,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			run(cmd, engine.NewStatusCommand(workFlag, rootFlag))
			return nil
		},
	}
	statusCmd.Flags().StringVar(&workFlag, "work", "", "Select a work item by immutable ID")

	var workflowFlag string
	var dryRunFlag bool
	startCmd := &cobra.Command{
		Use:   "start --dry-run \"intent\"",
		Short: "Preview a work-creation plan without applying it",
		Long:  help.StartHelp,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !dryRunFlag {
				*exit = renderOutcomeFor(stdout, stderr, eng, jsonFlag, outputFlag, env,
					eng.UsageError(engine.Command{}, "only --dry-run preview is implemented; durable start arrives with workspace transactions"))
				return nil
			}
			if strings.TrimSpace(args[0]) == "" {
				*exit = renderOutcomeFor(stdout, stderr, eng, jsonFlag, outputFlag, env,
					eng.UsageError(engine.Command{}, "start intent is required"))
				return nil
			}
			start, err := engine.NewStartDryRunCommand(args[0], workflowFlag, rootFlag)
			if err != nil {
				panic(fmt.Sprintf("cli: validated start command rejected: %v", err))
			}
			run(cmd, start)
			return nil
		},
	}
	startCmd.Flags().StringVar(&workflowFlag, "workflow", "", "Journey the new work item belongs to")
	startCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Preview the creation plan without applying it")

	helpCmd := &cobra.Command{
		Use:   "help [topic]",
		Short: "Show command help without touching a workspace",
		Long:  help.Text,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			topic := ""
			if len(args) == 1 {
				topic = args[0]
			}
			*exit = printHelp(cmd.OutOrStdout(), cmd.ErrOrStderr(), topic)
			return nil
		},
	}
	root.AddCommand(versionCmd, doctorCmd, statusCmd, startCmd, helpCmd)
	return root
}

// cliFailure translates a Cobra dispatch failure into a stable outcome.
// Cobra only fails before a subcommand runs (unknown input, bad flags,
// arity violations), so the failure is always a usage error. An unknown
// bare token preserves its name wherever flag placement puts it; anything
// else normalizes to an unspecified usage failure rather than trusting
// unparsed input.
func cliFailure(stdout, stderr io.Writer, argv []string, env Env, eng *engine.Engine) int {
	if name := unknownName(argv); name != "" {
		outcome := eng.UnknownCommand(name)
		return renderOutcomeFor(stdout, stderr, eng, jsonFlagOf(argv), outputFlagOf(argv), env, outcome)
	}
	return renderOutcomeFor(stdout, stderr, eng, jsonFlagOf(argv), outputFlagOf(argv), env,
		eng.UsageError(engine.Command{}, "invalid command input"))
}

// unknownName extracts the unknown command token, skipping known flags and
// their values so leading options cannot hide it. Once an unknown flag is
// observed, later bare tokens are ambiguous flag values and are never copied
// into the outcome. This fail-closed rule prevents credentials supplied to a
// misspelled or unsupported option from leaking as requestedCommand.
func unknownName(argv []string) string {
	valueFlags := map[string]bool{
		"--output": true, "--root": true, "--work": true, "--workflow": true,
	}
	boolFlags := map[string]bool{
		"--json": true, "--dry-run": true, "-h": true, "--help": true, "--version": true,
	}
	tokens := preFlagTokens(argv)
	unknownFlag := false
	for index := 0; index < len(tokens); index++ {
		token := tokens[index]
		if !strings.HasPrefix(token, "-") || token == "-" {
			if unknownFlag {
				continue
			}
			if !isKnownCommand(token) {
				return token
			}
			continue
		}
		if boolFlags[token] {
			continue
		}
		// Any --name=value form is an option, known or not.
		if strings.Contains(token, "=") {
			continue
		}
		if valueFlags[token] {
			index++
			continue
		}
		unknownFlag = true
	}
	return ""
}

func isKnownCommand(name string) bool {
	switch name {
	case "version", "doctor", "status", "start", "help":
		return true
	default:
		return false
	}
}

func jsonFlagOf(argv []string) bool {
	for _, token := range preFlagTokens(argv) {
		switch token {
		case "--json", "--json=true":
			return true
		}
	}
	return false
}

func outputFlagOf(argv []string) string {
	tokens := preFlagTokens(argv)
	for index, token := range tokens {
		if token == "--output" && index+1 < len(tokens) {
			return tokens[index+1]
		}
		if value, ok := strings.CutPrefix(token, "--output="); ok {
			return value
		}
	}
	return ""
}

// renderOutcomeFor resolves the presentation mode and renders one outcome,
// returning the process exit code. It is the single translation point from
// outcome to streams and exits. Option failures become usage outcomes and
// render through the same path, so a JSON caller always receives a
// parseable envelope on stdout instead of a bare stderr line.
func renderOutcomeFor(
	stdout, stderr io.Writer,
	eng *engine.Engine,
	jsonFlag bool,
	outputFlag string,
	env Env,
	outcome engine.Outcome,
) int {
	if jsonFlag && outputFlag != "" && outputFlag != "json" {
		// The raw flag value is deliberately not echoed: option values
		// are untrusted input and have no place in a diagnostic.
		return renderOutcomeFor(stdout, stderr, eng, jsonFlag, "", env,
			eng.UsageError(engine.Command{}, "conflicting output options: --json cannot combine with --output"))
	}
	explicit := outputFlag
	if explicit == "" && jsonFlag {
		explicit = "json"
	}
	mode, err := outputmode.Resolve(outputmode.Request{
		ExplicitOutput: explicit,
		MachineCaller:  explicit == "json",
		StdinTerminal:  env.StdinTerminal,
		StdoutTerminal: env.StdoutTerminal,
	})
	if err != nil {
		return renderOutcomeFor(stdout, stderr, eng, jsonFlag, "", env,
			eng.UsageError(engine.Command{}, err.Error()))
	}
	input := render.NewInput(outcome, render.NewRedactor())
	switch mode {
	case outputmode.ModeJSON:
		if err := render.WriteJSON(stdout, input); err != nil {
			fmt.Fprintln(stderr, "render output write failed")
			return 1
		}
		return int(outcome.ExitCode())
	case outputmode.ModeTUI:
		return renderOutcomeFor(stdout, stderr, eng, jsonFlag, "", env,
			eng.UsageError(engine.Command{}, "interactive terminal output is not implemented in foundation"))
	default:
		if err := render.WriteHuman(stdout, stderr, input); err != nil {
			fmt.Fprintln(stderr, "render output write failed")
			return 1
		}
		return int(outcome.ExitCode())
	}
}
