package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSingleRecord(t *testing.T) {
	dir := t.TempDir()
	if err := run([]string{"--dir", dir}, strings.NewReader("web,7\n"), io.Discard); err != nil {
		t.Fatal(err)
	}
	assertRecords(t, dir, map[string]string{"web": "7\n"})
}

func TestRunBatchSuccess(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input string
		want  map[string]string
	}{
		{"empty", "", nil},
		{"ordered records", "web,7\napi,12\nworker,3", map[string]string{"web": "7\n", "api": "12\n", "worker": "3\n"}},
		{"duplicate replaces", "web,7\napi,12\nweb,9\n", map[string]string{"web": "9\n", "api": "12\n"}},
		{"CSV and normalization", "\"web\",0007\r\n\"two,words\",+012\r\n", map[string]string{"web": "7\n", "two,words": "12\n"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "new", "records")
			if err := run([]string{"--dir", dir}, strings.NewReader(tc.input), io.Discard); err != nil {
				t.Fatal(err)
			}
			assertRecords(t, dir, tc.want)
		})
	}
}

func TestRunStopsAtInvalidRow(t *testing.T) {
	for _, tc := range []struct {
		name string
		row  string
	}{
		{"too few fields", "bad"},
		{"too many fields", "bad,1,extra"},
		{"empty ID", ",1"},
		{"dot ID", ".,1"},
		{"dot dot ID", "..,1"},
		{"slash ID", "nested/bad,1"},
		{"backslash ID", "nested\\bad,1"},
		{"empty quantity", "bad,"},
		{"zero quantity", "bad,0"},
		{"negative duplicate", "keep,-9"},
		{"fractional quantity", "bad,1.5"},
		{"nondecimal quantity", "bad,0x10"},
		{"space in quantity", "bad, 1"},
		{"overflow quantity", "bad,999999999999999999999999999999"},
		{"malformed CSV", "bad,\"unterminated"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFixture(t, filepath.Join(dir, "keep"), "11\n")
			input := "first,3\nkeep,5\n" + tc.row + "\nlater,7\n"
			if err := run([]string{"--dir", dir}, strings.NewReader(input), io.Discard); err == nil {
				t.Fatal("run succeeded with invalid input")
			}
			assertRecords(t, dir, map[string]string{"first": "3\n", "keep": "5\n"})
		})
	}
}

func TestRunRejectsFirstRowWithoutWrites(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "records")
	if err := run([]string{"--dir", dir}, strings.NewReader("bad,0\nlater,7\n"), io.Discard); err == nil {
		t.Fatal("run succeeded with invalid first row")
	}
	assertRecords(t, dir, nil)
}

func TestRunStopsAtFileFailure(t *testing.T) {
	dir := t.TempDir()
	blocked := filepath.Join(dir, "blocked")
	if err := os.Mkdir(blocked, 0o755); err != nil {
		t.Fatal(err)
	}
	input := "first,3\nfirst,5\nblocked,9\nlater,7\n"
	if err := run([]string{"--dir", dir}, strings.NewReader(input), io.Discard); err == nil {
		t.Fatal("run succeeded writing over a directory")
	}
	assertFile(t, filepath.Join(dir, "first"), "5\n")
	assertNames(t, dir, []string{"blocked", "first"})
	info, err := os.Stat(blocked)
	if err != nil || !info.IsDir() {
		t.Fatalf("blocked path = %v, %v; want unchanged directory", info, err)
	}
}

func TestRunDirectoryFailure(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "records")
	writeFixture(t, dir, "existing file\n")
	if err := run([]string{"--dir", dir}, strings.NewReader("first,3\nlater,7\n"), io.Discard); err == nil {
		t.Fatal("run succeeded with a file as its directory")
	}
	assertFile(t, dir, "existing file\n")
}

func TestRunInputFailureRetainsPrefix(t *testing.T) {
	dir := t.TempDir()
	inputErr := errors.New("input interrupted")
	input := io.MultiReader(strings.NewReader("first,3\nkeep,5\n"), errorReader{inputErr})
	err := run([]string{"--dir", dir}, input, io.Discard)
	if err == nil || !strings.Contains(err.Error(), inputErr.Error()) {
		t.Fatalf("run error = %v, want useful input diagnostic", err)
	}
	assertRecords(t, dir, map[string]string{"first": "3\n", "keep": "5\n"})
}

func TestRunPublishesBeforeReadingNextRecord(t *testing.T) {
	dir := t.TempDir()
	observed := false
	input := &observedReader{
		prefix: strings.NewReader("first,3\n"),
		suffix: strings.NewReader("second,7\n"),
		observe: func() error {
			observed = true
			got, err := os.ReadFile(filepath.Join(dir, "first"))
			if err != nil || string(got) != "3\n" {
				return fmt.Errorf("first record not published before next read: %q, %v", got, err)
			}
			return nil
		},
	}
	if err := run([]string{"--dir", dir}, input, io.Discard); err != nil {
		t.Fatal(err)
	}
	if !observed {
		t.Fatal("did not read the second record")
	}
	assertRecords(t, dir, map[string]string{"first": "3\n", "second": "7\n"})
}

func TestCLIProcessContract(t *testing.T) {
	binary := buildCLI(t)
	for _, tc := range []struct {
		name       string
		input      string
		args       func(string) []string
		code       int
		diagnostic string
		want       map[string]string
	}{
		{"empty", "", dirArgs, 0, "", nil},
		{"single record", "web,7\n", dirArgs, 0, "", map[string]string{"web": "7\n"}},
		{"batch and duplicate", "web,7\napi,12\nweb,9\n", dirArgs, 0, "", map[string]string{"web": "9\n", "api": "12\n"}},
		{"invalid first row", "bad,0\nlater,7\n", dirArgs, 2, "quantity", nil},
		{"invalid duplicate", "keep,5\nkeep,-9\nlater,7\n", dirArgs, 2, "quantity", map[string]string{"keep": "5\n"}},
		{"malformed CSV after prefix", "keep,5\nbad,\"unterminated\nlater,7\n", dirArgs, 2, "quote", map[string]string{"keep": "5\n"}},
		{"missing directory", "web,7\n", func(string) []string { return nil }, 2, "dir", nil},
		{"empty directory option", "web,7\n", func(string) []string { return []string{"--dir", ""} }, 2, "dir", nil},
		{"positional argument", "web,7\n", func(dir string) []string { return []string{"--dir", dir, "extra"} }, 2, "positional", nil},
		{"unknown flag", "web,7\n", func(string) []string { return []string{"--unknown"} }, 2, "unknown", nil},
		{"missing flag value", "web,7\n", func(string) []string { return []string{"--dir"} }, 2, "argument", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "records")
			code, stdout, stderr := runCLI(t, binary, tc.args(dir), tc.input)
			assertProcess(t, code, stdout, stderr, tc.code, tc.diagnostic)
			assertRecords(t, dir, tc.want)
		})
	}
	t.Run("file failure after prefix", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Mkdir(filepath.Join(dir, "blocked"), 0o755); err != nil {
			t.Fatal(err)
		}
		code, stdout, stderr := runCLI(t, binary, dirArgs(dir), "first,3\nblocked,9\nlater,7\n")
		assertProcess(t, code, stdout, stderr, 2, "blocked")
		assertFile(t, filepath.Join(dir, "first"), "3\n")
		assertNames(t, dir, []string{"blocked", "first"})
	})
	t.Run("directory creation failure", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "records")
		writeFixture(t, dir, "existing file\n")
		code, stdout, stderr := runCLI(t, binary, dirArgs(dir), "first,3\nlater,7\n")
		assertProcess(t, code, stdout, stderr, 2, "records")
		assertFile(t, dir, "existing file\n")
	})
}

type errorReader struct{ err error }

func (r errorReader) Read([]byte) (int, error) { return 0, r.err }

type observedReader struct {
	prefix  *strings.Reader
	suffix  *strings.Reader
	observe func() error
}

func (r *observedReader) Read(p []byte) (int, error) {
	if r.prefix.Len() > 0 {
		return r.prefix.Read(p)
	}
	if r.observe != nil {
		observe := r.observe
		r.observe = nil
		if err := observe(); err != nil {
			return 0, err
		}
	}
	return r.suffix.Read(p)
}

func dirArgs(dir string) []string { return []string{"--dir", dir} }

func writeFixture(t *testing.T, path, value string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != want {
		t.Fatalf("%s = %q, %v; want %q", path, got, err, want)
	}
}

func assertRecords(t *testing.T, dir string, want map[string]string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) && len(want) == 0 {
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != len(want) {
		t.Fatalf("entries = %v, want records %v", entries, want)
	}
	for _, entry := range entries {
		value, ok := want[entry.Name()]
		if !ok {
			t.Fatalf("unexpected entry %q", entry.Name())
		}
		assertFile(t, filepath.Join(dir, entry.Name()), value)
	}
}

func assertNames(t *testing.T, dir string, want []string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	got := make([]string, len(entries))
	for i, entry := range entries {
		got[i] = entry.Name()
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("entries = %v, want %v", got, want)
	}
}

func buildCLI(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "ledgerload")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOCACHE=/private/tmp/go-quality-testing-cache")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build executable: %v (%v)\n%s", err, ctx.Err(), output)
	}
	return binary
}

func runCLI(t *testing.T, binary string, args []string, input string) (int, string, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("executable exceeded deadline: %v", ctx.Err())
	}
	if err == nil {
		return 0, stdout.String(), stderr.String()
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("start executable: %v", err)
	}
	return exitErr.ExitCode(), stdout.String(), stderr.String()
}

func assertProcess(t *testing.T, code int, stdout, stderr string, wantCode int, diagnostic string) {
	t.Helper()
	if code != wantCode || stdout != "" {
		t.Fatalf("process = exit %d, stdout %q, stderr %q; want exit %d and empty stdout", code, stdout, stderr, wantCode)
	}
	if wantCode == 0 && stderr != "" {
		t.Fatalf("success stderr = %q, want empty", stderr)
	}
	if wantCode != 0 && (strings.TrimSpace(stderr) == "" || !strings.Contains(stderr, diagnostic)) {
		t.Fatalf("failure stderr = %q, want useful %q diagnostic", stderr, diagnostic)
	}
}
