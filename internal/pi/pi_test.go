package pi

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAskBuildsCorrectArgs(t *testing.T) {
	t.Setenv("BOBOT_TOOLS", "") // default: persona-only -> --no-tools
	var got []string
	stub := func(_ context.Context, args ...string) *exec.Cmd {
		got = append(got, args...)
		// Fake a successful pi run that echoes a fixed reply to stdout.
		return exec.Command("sh", "-c", "echo 'hi back'")
	}
	out, err := run(context.Background(), stub, "chan-123", "hello")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if got := strings.TrimSpace(out); got != "hi back" {
		t.Fatalf("output=%q want %q", out, "hi back")
	}

	wantFlags := map[string]bool{
		"-p": true, "--no-tools": true, "--no-context-files": true,
		"--session-id": true, "--session-dir": true, "--system-prompt": true,
	}
	for flag := range wantFlags {
		if !contains(got, flag) {
			t.Fatalf("missing flag %q in %v", flag, got)
		}
	}

	// System prompt value is present right after --system-prompt.
	i := indexOf(got, "--system-prompt")
	if i < 0 || !strings.Contains(got[i+1], "bobot") {
		t.Fatalf("system prompt missing/invalid: %v", got)
	}

	// Session id value is `chan-123`, right after --session-id.
	if i = indexOf(got, "--session-id"); i < 0 || got[i+1] != "chan-123" {
		t.Fatalf("session-id missing/invalid: %v", got)
	}

	// Session dir is the CWD-relative SessionDir constant.
	if i = indexOf(got, "--session-dir"); i < 0 || got[i+1] != SessionDir {
		t.Fatalf("session-dir missing/invalid: %v", got)
	}

	// User prompt is last.
	if got[len(got)-1] != "hello" {
		t.Fatalf("last arg=%q want %q", got[len(got)-1], "hello")
	}
}

func TestAskPropagatesFailure(t *testing.T) {
	stub := func(_ context.Context, args ...string) *exec.Cmd {
		return exec.Command("sh", "-c", "exit 3")
	}
	if _, err := run(context.Background(), stub, "chan", "x"); err == nil {
		t.Fatal("expected error on non-zero exit")
	}
}

func TestAskTimeoutHonorsDeadline(t *testing.T) {
	orig := buildCmd
	buildCmd = func(ctx context.Context, args ...string) *exec.Cmd {
		// Sleep longer than the timeout; Cmd must honor done ctx and kill.
		return exec.CommandContext(ctx, "sh", "-c", "sleep 10")
	}
	defer func() { buildCmd = orig }()

	if _, err := AskTimeout("chan", "hi", 20*time.Millisecond); err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestAskTrimsWhitespace(t *testing.T) {
	orig := buildCmd
	buildCmd = func(ctx context.Context, args ...string) *exec.Cmd {
		return exec.Command("sh", "-c", "printf '\\n  padded  \\n'")
	}
	defer func() { buildCmd = orig }()

	got, err := Ask(context.Background(), "chan", "x")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if got != "padded" {
		t.Fatalf("Ask=%q want %q", got, "padded")
	}
}

func TestClearSessionRemovesMatchedFiles(t *testing.T) {
	orig := SessionDir
	dir := t.TempDir()
	SessionDir = dir
	defer func() { SessionDir = orig }()

	target := filepath.Join(dir, "2026-01-01T00-00-00-000Z_chan-1.jsonl")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// unrelated session for a different channel must be left intact.
	other := filepath.Join(dir, "2026-01-01T00-00-00-000Z_chan-2.jsonl")
	if err := os.WriteFile(other, []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := ClearSession("chan-1"); err != nil {
		t.Fatalf("ClearSession: %v", err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("target still exists: %v", err)
	}
	if _, err := os.Stat(other); err != nil {
		t.Fatalf("other removed: %v", err)
	}
}

func TestClearSessionNoSessionIsNoop(t *testing.T) {
	orig := SessionDir
	SessionDir = t.TempDir()
	defer func() { SessionDir = orig }()

	if err := ClearSession("nonexistent"); err != nil {
		t.Fatalf("expected nil on missing session, got %v", err)
	}
}

func TestToolFlagsNoToolsByDefault(t *testing.T) {
	t.Setenv("BOBOT_TOOLS", "")
	flags := toolFlags()
	if len(flags) != 1 || flags[0] != "--no-tools" {
		t.Fatalf("default want [--no-tools], got %v", flags)
	}
}

func TestToolFlagsAllowlist(t *testing.T) {
	t.Setenv("BOBOT_TOOLS", "brave_search,web")
	flags := toolFlags()
	if len(flags) != 2 || flags[0] != "--tools" || flags[1] != "brave_search,web" {
		t.Fatalf("allowlist want [--tools brave_search,web], got %v", flags)
	}
}

func TestToolFlagsAll(t *testing.T) {
	t.Setenv("BOBOT_TOOLS", "all")
	if flags := toolFlags(); flags != nil {
		t.Fatalf("all want nil (no flag), got %v", flags)
	}
}

// TestAskStderrExcludedFromReply guards the regression where pi's stderr
// diagnostics (e.g. "Warning: No project session found...") leaked into the
// Discord reply because stdout and stderr shared a buffer. Reply = stdout only.
func TestAskStderrExcludedFromReply(t *testing.T) {
	orig := buildCmd
	buildCmd = func(ctx context.Context, args ...string) *exec.Cmd {
		// reply on stdout, diagnostic noise on stderr
		return exec.Command("sh", "-c", "echo 'Warning: No project session found' >&2; echo 'hi back'")
	}
	defer func() { buildCmd = orig }()

	got, err := Ask(context.Background(), "chan", "x")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if got != "hi back" {
		t.Fatalf("reply=%q want %q (stderr leaked)", got, "hi back")
	}
}

// guard against stdin pipe hangs on some pi builds
var _ = bytes.NewReader

func contains(s []string, v string) bool { return indexOf(s, v) >= 0 }

func indexOf(s []string, v string) int {
	for i, x := range s {
		if x == v {
			return i
		}
	}
	return -1
}