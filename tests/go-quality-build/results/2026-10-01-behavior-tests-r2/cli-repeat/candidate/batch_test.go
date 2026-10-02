package main

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func assertFiles(t *testing.T, dir string, want map[string]string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != len(want) {
		t.Fatalf("directory entries = %v, want %d files", entries, len(want))
	}
	for _, entry := range entries {
		content, ok := want[entry.Name()]
		if !ok {
			t.Errorf("unexpected entry %q", entry.Name())
			continue
		}
		got, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil || string(got) != content {
			t.Errorf("%s = %q, %v; want %q", entry.Name(), got, err, content)
		}
	}
}

func TestBatchSuccess(t *testing.T) {
	dir := t.TempDir()
	input := "web,0007\nworker,2\nweb,+009\n\"with,comma\",4\n"
	if err := run([]string{"--dir", dir}, strings.NewReader(input), io.Discard); err != nil {
		t.Fatal(err)
	}
	assertFiles(t, dir, map[string]string{"web": "9\n", "worker": "2\n", "with,comma": "4\n"})
}

func TestBatchStopsAtInvalidReplacement(t *testing.T) {
	for _, tc := range []struct {
		name, row string
	}{
		{"empty ID", ",2"},
		{"dot", ".,2"},
		{"dot dot", "..,2"},
		{"slash", "a/b,2"},
		{"backslash", "a\\b,2"},
		{"zero", "web,0"},
		{"negative", "web,-1"},
		{"empty quantity", "web,"},
		{"hexadecimal", "web,0x10"},
		{"fraction", "web,1.5"},
		{"space", "web, 2"},
		{"overflow", "web,99999999999999999999999999999"},
		{"missing field", "web"},
		{"extra field", "web,2,extra"},
		{"malformed quote", "web,\"2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			input := "web,7\nother,3\n" + tc.row + "\nlater,6\n"
			err := run([]string{"--dir", dir}, strings.NewReader(input), io.Discard)
			if err == nil || err.Error() == "" {
				t.Fatalf("run error = %v, want diagnostic", err)
			}
			assertFiles(t, dir, map[string]string{"web": "7\n", "other": "3\n"})
		})
	}
}

func TestFirstInvalidRowDoesNotWrite(t *testing.T) {
	dir := t.TempDir()
	if err := run([]string{"--dir", dir}, strings.NewReader("bad,0\nlater,1\n"), io.Discard); err == nil {
		t.Fatal("invalid first row succeeded")
	}
	assertFiles(t, dir, map[string]string{})
}

type readerFunc func([]byte) (int, error)

func (f readerFunc) Read(p []byte) (int, error) { return f(p) }

func TestWritesBeforeReadingNextRow(t *testing.T) {
	dir := t.TempDir()
	read := 0
	input := readerFunc(func(p []byte) (int, error) {
		read++
		switch read {
		case 1:
			return copy(p, "first,3\n"), nil
		case 2:
			assertFiles(t, dir, map[string]string{"first": "3\n"})
			return copy(p, "second,4\n"), nil
		default:
			return 0, io.EOF
		}
	})
	if err := run([]string{"--dir", dir}, input, io.Discard); err != nil {
		t.Fatal(err)
	}
	assertFiles(t, dir, map[string]string{"first": "3\n", "second": "4\n"})
}

func TestReadFailureRetainsPrefix(t *testing.T) {
	dir := t.TempDir()
	readErr := errors.New("input unavailable")
	read := false
	input := readerFunc(func(p []byte) (int, error) {
		if read {
			return 0, readErr
		}
		read = true
		return copy(p, "first,3\n"), nil
	})
	if err := run([]string{"--dir", dir}, input, io.Discard); err == nil {
		t.Fatal("input read failure succeeded")
	}
	assertFiles(t, dir, map[string]string{"first": "3\n"})
}

func TestFileFailureRetainsPrefix(t *testing.T) {
	dir := t.TempDir()
	blocked := filepath.Join(dir, "blocked")
	if err := os.Mkdir(blocked, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(blocked, "marker"), []byte("unchanged"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := run([]string{"--dir", dir}, strings.NewReader("first,3\nblocked,4\nlater,5\n"), io.Discard)
	if err == nil {
		t.Fatal("blocked file succeeded")
	}
	got, err := os.ReadFile(filepath.Join(dir, "first"))
	if err != nil || string(got) != "3\n" {
		t.Fatalf("accepted file = %q, %v", got, err)
	}
	assertFiles(t, blocked, map[string]string{"marker": "unchanged"})
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 2 {
		t.Fatalf("directory after failure = %v, %v; want first and blocked only", entries, err)
	}
}

func buildCLI(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "ledgerload")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, out)
	}
	return binary
}

func checkProcess(t *testing.T, cmd *exec.Cmd, input string, wantCode int) string {
	t.Helper()
	var stdout, stderr strings.Builder
	cmd.Stdin = strings.NewReader(input)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("execute CLI: %v", err)
		}
		code = exitErr.ExitCode()
	}
	if code != wantCode || stdout.Len() != 0 || (wantCode == 0 && stderr.Len() != 0) || (wantCode != 0 && stderr.Len() == 0) {
		t.Fatalf("CLI code=%d stdout=%q stderr=%q; want code=%d, empty stdout, diagnostic only on failure", code, stdout.String(), stderr.String(), wantCode)
	}
	return stderr.String()
}

func TestCLIContract(t *testing.T) {
	binary := buildCLI(t)
	for _, tc := range []struct {
		name, input, diagnostic string
		args                    []string
		code                    int
		files                   map[string]string
	}{
		{name: "single record", input: "web,7\n", code: 0, files: map[string]string{"web": "7\n"}},
		{name: "batch and duplicate", input: "web,7\nworker,2\nweb,9\n", code: 0, files: map[string]string{"web": "9\n", "worker": "2\n"}},
		{name: "empty input", code: 0, files: map[string]string{}},
		{name: "invalid first row", input: "web,0\nlater,4\n", diagnostic: "quantity", code: 2, files: map[string]string{}},
		{name: "invalid replacement", input: "web,7\nweb,0\nlater,4\n", diagnostic: "quantity", code: 2, files: map[string]string{"web": "7\n"}},
		{name: "malformed CSV", input: "web,7\nworker,\"2\nlater,4\n", diagnostic: "read record 2", code: 2, files: map[string]string{"web": "7\n"}},
		{name: "missing dir", args: []string{}, input: "web,7\n", diagnostic: "--dir", code: 2, files: map[string]string{}},
		{name: "empty dir", args: []string{"--dir="}, input: "web,7\n", diagnostic: "--dir", code: 2, files: map[string]string{}},
		{name: "missing flag value", args: []string{"--dir"}, input: "web,7\n", diagnostic: "dir", code: 2, files: map[string]string{}},
		{name: "positional argument", args: []string{"--dir", "DIR", "extra"}, input: "web,7\n", diagnostic: "positional", code: 2, files: map[string]string{}},
		{name: "unknown flag", args: []string{"--unknown"}, input: "web,7\n", diagnostic: "unknown", code: 2, files: map[string]string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			args := []string{"--dir", dir}
			if tc.args != nil {
				args = append([]string{}, tc.args...)
				for i := range args {
					if args[i] == "DIR" {
						args[i] = dir
					}
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			diagnostic := checkProcess(t, exec.CommandContext(ctx, binary, args...), tc.input, tc.code)
			if tc.diagnostic != "" && !strings.Contains(diagnostic, tc.diagnostic) {
				t.Errorf("diagnostic %q does not identify %q", diagnostic, tc.diagnostic)
			}
			assertFiles(t, dir, tc.files)
		})
	}
	t.Run("directory creation", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "new", "records")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		checkProcess(t, exec.CommandContext(ctx, binary, "--dir", dir), "web,7\n", 0)
		assertFiles(t, dir, map[string]string{"web": "7\n"})
	})
	t.Run("directory failure", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "file")
		if err := os.WriteFile(dir, []byte("unchanged"), 0o644); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		checkProcess(t, exec.CommandContext(ctx, binary, "--dir", dir), "web,7\nlater,4\n", 2)
		got, err := os.ReadFile(dir)
		if err != nil || string(got) != "unchanged" {
			t.Fatalf("directory obstruction = %q, %v", got, err)
		}
	})
	t.Run("file failure", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Mkdir(filepath.Join(dir, "blocked"), 0o755); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		checkProcess(t, exec.CommandContext(ctx, binary, "--dir", dir), "web,7\nblocked,2\nlater,4\n", 2)
		got, err := os.ReadFile(filepath.Join(dir, "web"))
		if err != nil || string(got) != "7\n" {
			t.Fatalf("accepted file = %q, %v", got, err)
		}
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) != 2 {
			t.Fatalf("directory = %v, %v; want accepted file and obstruction only", entries, err)
		}
	})
}
