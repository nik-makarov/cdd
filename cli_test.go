package main

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// Keep flag as a test oracle only: compatibility must not cost binary size.
func standardArgs(args []string) (options, error) {
	var opts options
	flags := flag.NewFlagSet("cdd", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.BoolVar(&opts.hidden, "hidden", false, "")
	flags.BoolVar(&opts.showVersion, "version", false, "")
	flags.StringVar(&opts.initShell, "init", "", "")
	flags.Usage = func() {}
	err := flags.Parse(args)
	if errors.Is(err, flag.ErrHelp) {
		opts.showHelp = true
		err = nil
	}
	opts.args = flags.Args()
	return opts, err
}

func checkArgumentCompatibility(t *testing.T, args []string) {
	t.Helper()
	got, gotErr := parseArgs(args)
	want, wantErr := standardArgs(args)
	if (gotErr == nil) != (wantErr == nil) {
		t.Fatalf("args %q: got error %v, want %v", args, gotErr, wantErr)
	}
	if gotErr != nil {
		if gotErr.Error() != wantErr.Error() {
			t.Fatalf("args %q: got error %q, want %q", args, gotErr, wantErr)
		}
		return
	}
	if got.hidden != want.hidden || got.showVersion != want.showVersion || got.showHelp != want.showHelp || got.initShell != want.initShell || (!got.showHelp && !slices.Equal(got.args, want.args)) {
		t.Fatalf("args %q: got %+v, want %+v", args, got, want)
	}
}

func TestParseArgsCompatibility(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
	}{
		{"defaults", nil},
		{"hidden", []string{"--hidden", "projects"}},
		{"single dash", []string{"-hidden", "-init", "zsh", "-version"}},
		{"equals", []string{"--init=bash", "--hidden=true", "--version=1"}},
		{"last value wins", []string{"--hidden", "--hidden=false", "--init=zsh", "--init=bash"}},
		{"terminator", []string{"--", "--hidden"}},
		{"first positional", []string{"projects", "--hidden"}},
		{"literal dash", []string{"-", "--hidden"}},
		{"empty path", []string{"", "--version"}},
		{"help", []string{"--help", "--unknown"}},
		{"help value", []string{"-h=false", "a", "b"}},
		{"long help alias", []string{"--h"}},
		{"short help alias", []string{"-help"}},
		{"string consumes flag", []string{"--init", "--help"}},
		{"string consumes terminator", []string{"--init", "--"}},
		{"empty string", []string{"--init="}},
		{"empty string argument", []string{"--init", ""}},
		{"equals in value", []string{"--init=zsh=value"}},
		{"missing value", []string{"--init"}},
		{"invalid bool", []string{"--hidden=maybe"}},
		{"empty bool", []string{"--version="}},
		{"separate bool is positional", []string{"--hidden", "false"}},
		{"unknown option", []string{"--unknown=value"}},
		{"unknown before help", []string{"--unknown", "--help"}},
		{"malformed option", []string{"---hidden"}},
		{"malformed assignment", []string{"--=hidden"}},
	} {
		t.Run(tt.name, func(t *testing.T) { checkArgumentCompatibility(t, tt.args) })
	}
	for _, value := range []string{"1", "t", "T", "TRUE", "true", "True", "0", "f", "F", "FALSE", "false", "False"} {
		t.Run("bool/"+value, func(t *testing.T) {
			checkArgumentCompatibility(t, []string{"--hidden=" + value, "--version=" + value})
		})
	}
}

func FuzzParseArgsCompatibility(f *testing.F) {
	for _, seed := range []string{"--hidden\x00projects", "--init=zsh", "--help\x00--unknown", "---bad", "--hidden=invalid", "--\x00--version"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		checkArgumentCompatibility(t, strings.Split(input, "\x00"))
	})
}

func TestCLIWithoutTerminal(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
		code int
	}{
		{"help", []string{"--help"}, 0},
		{"version", []string{"--version"}, 0},
		{"init zsh", []string{"--init", "zsh"}, 0},
		{"init bash", []string{"--init", "bash"}, 0},
		{"unknown shell", []string{"--init", "fish"}, 2},
		{"unknown flag", []string{"--unknown"}, 2},
		{"missing init value", []string{"--init"}, 2},
		{"invalid bool", []string{"--hidden=maybe"}, 2},
		{"help before invalid args", []string{"--help", "a", "b"}, 0},
		{"positional before help", []string{"a", "--help"}, 2},
		{"version equals", []string{"--version=true"}, 0},
		{"extra args", []string{"a", "b"}, 2},
		{"missing directory", []string{filepath.Join(t.TempDir(), "missing")}, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			if got := run(tt.args, &out, &errOut); got != tt.code {
				t.Fatalf("exit = %d, want %d; stderr: %s", got, tt.code, &errOut)
			}
			if tt.code != 0 && out.Len() != 0 {
				t.Fatalf("error contaminated stdout: %q", &out)
			}
			if tt.name == "help" && (!strings.Contains(errOut.String(), "Usage:") || out.Len() != 0) {
				t.Fatal("help missing or leaked into stdout")
			}
			if tt.name == "version" && out.String() != "cdd "+version+"\n" {
				t.Fatalf("version output: %q", &out)
			}
		})
	}
}

func TestWriteResultFailure(t *testing.T) {
	var stderr bytes.Buffer
	if code := writeResult(failingWriter{}, &stderr, "chosen path"); code != 1 || !strings.Contains(stderr.String(), "cdd: write output:") {
		t.Fatalf("write failure not reported: exit %d, stderr %q", code, stderr.String())
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }
