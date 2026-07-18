package pi

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestAskBuildsCorrectArgs(t *testing.T) {
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