// Package pi spawns pi coding-agent instances to answer user messages.
package pi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// systemPrompt is the assistant persona pi adopts for Discord replies.
const systemPrompt = "You are bobot, a concise assistant in a Discord chat. " +
	"Reply briefly and conversationally. No code blocks unless asked."

// SessionDir is CWD-relative so bobot's sessions stay isolated from the
// operator's own global pi session store; cleanup is `rm -rf ./sessions`.
// var (not const) so tests can point it at a temp dir.
var SessionDir = "./sessions"

// Ask runs a non-interactive pi instance scoped to sessionID, so each Discord
// channel/thread carries its own conversation memory across turns.
func Ask(ctx context.Context, sessionID, prompt string) (string, error) {
	out, err := run(ctx, buildCmd, sessionID, prompt)
	if err != nil {
		return "", fmt.Errorf("pi: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// AskTimeout is Ask with a bounded deadline.
func AskTimeout(sessionID, prompt string, d time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	return Ask(ctx, sessionID, prompt)
}

// ClearSession deletes the on-disk session file for sessionID if any.
// Matches pi's layout `<session-dir>/<timestamp>_<session-id>.jsonl`.
// ponytail: glob-and-remove is safe because our sessionID is a Discord
// channel ID (numeric snowflake, no glob meta); escape if session keys ever
// carry shell/glob characters.
func ClearSession(sessionID string) error {
	matches, err := filepath.Glob(filepath.Join(SessionDir, "*_"+sessionID+".jsonl"))
	if err != nil {
		return fmt.Errorf("glob sessions: %w", err)
	}
	if len(matches) == 0 {
		return nil
	}
	for _, p := range matches {
		if err := os.Remove(p); err != nil {
			return fmt.Errorf("remove %s: %w", p, err)
		}
	}
	return nil
}

// commandFunc builds the exec.Cmd for a prompt; overridable in tests.
type commandFunc func(ctx context.Context, args ...string) *exec.Cmd

var buildCmd commandFunc = func(ctx context.Context, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, "pi", args...)
}

// run executes buildCmd for prompt scoped to sessionID and returns its stdout.
func run(ctx context.Context, build commandFunc, sessionID, prompt string) (string, error) {
	// ponytail: no-tools/no-context-files keep this a pure chat turn (no file
	// access, no project AGENTS.md leaking). --session-id + --session-dir give
	// per-channel memory isolated from the operator's own pi sessions; pi does
	// create-if-missing/continue-if-exists for free here. Upgrade to GC if
	// ./sessions growth ever matters.
	args := []string{
		"-p",
		"--no-tools",
		"--no-context-files",
		"--session-dir", SessionDir,
		"--session-id", sessionID,
		"--system-prompt", systemPrompt,
		prompt,
	}
	cmd := build(ctx, args...)
	cmd.Stdin = io.NopCloser(bytes.NewReader(nil))

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("run %q: %w", strings.Join(args, " "), err)
	}
	return out.String(), nil
}