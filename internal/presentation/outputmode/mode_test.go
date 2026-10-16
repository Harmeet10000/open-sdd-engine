package outputmode

import "testing"

func TestResolve(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		request Request
		want    Mode
		wantErr bool
	}{
		{
			name: "explicit JSON overrides interactive terminal",
			request: Request{
				ExplicitOutput: "json",
				Interaction:    InteractionEnabled,
				StdinTerminal:  true,
				StdoutTerminal: true,
			},
			want: ModeJSON,
		},
		{
			name: "explicit text overrides machine caller",
			request: Request{
				ExplicitOutput: " text ",
				MachineCaller:  true,
			},
			want: ModeText,
		},
		{
			name: "explicit TUI selects a capable human terminal",
			request: Request{
				ExplicitOutput: "TUI",
				StdinTerminal:  true,
				StdoutTerminal: true,
			},
			want: ModeTUI,
		},
		{
			name: "explicit TUI cannot bypass machine safety",
			request: Request{
				ExplicitOutput: "TUI",
				MachineCaller:  true,
				StdinTerminal:  true,
				StdoutTerminal: true,
			},
			wantErr: true,
		},
		{
			name: "explicit TUI conflicts with non-interactive mode",
			request: Request{
				ExplicitOutput: "TUI",
				Interaction:    InteractionDisabled,
				StdinTerminal:  true,
				StdoutTerminal: true,
			},
			wantErr: true,
		},
		{
			name: "non-interactive human output is text",
			request: Request{
				Interaction: InteractionDisabled,
			},
			want: ModeText,
		},
		{
			name: "non-interactive machine output is JSON",
			request: Request{
				Interaction:   InteractionDisabled,
				MachineCaller: true,
			},
			want: ModeJSON,
		},
		{
			name: "interactive terminal can select TUI",
			request: Request{
				Interaction:    InteractionEnabled,
				StdinTerminal:  true,
				StdoutTerminal: true,
			},
			want: ModeTUI,
		},
		{
			name: "interactive redirected output is rejected",
			request: Request{
				Interaction:   InteractionEnabled,
				StdinTerminal: true,
			},
			wantErr: true,
		},
		{
			name: "interactive machine caller is rejected",
			request: Request{
				Interaction:    InteractionEnabled,
				MachineCaller:  true,
				StdinTerminal:  true,
				StdoutTerminal: true,
			},
			wantErr: true,
		},
		{
			name: "machine caller defaults to JSON",
			request: Request{
				MachineCaller: true,
			},
			want: ModeJSON,
		},
		{
			name: "redirected input defaults to text",
			request: Request{
				StdoutTerminal: true,
			},
			want: ModeText,
		},
		{
			name: "terminal heuristic remains concise text",
			request: Request{
				StdinTerminal:  true,
				StdoutTerminal: true,
			},
			want: ModeText,
		},
		{
			name: "unknown output fails closed",
			request: Request{
				ExplicitOutput: "xml",
			},
			wantErr: true,
		},
		{
			name: "unknown interaction fails closed",
			request: Request{
				Interaction: Interaction(99),
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := Resolve(test.request)
			if test.wantErr {
				if err == nil {
					t.Fatalf("Resolve(%+v) returned no error", test.request)
				}
				return
			}
			if err != nil {
				t.Fatalf("Resolve(%+v): %v", test.request, err)
			}
			if got != test.want {
				t.Fatalf("Resolve(%+v) = %q, want %q", test.request, got, test.want)
			}
		})
	}
}
