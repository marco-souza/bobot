// Package pi spawns pi coding-agent instances to answer user messages.
package pi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// BoBot opens extension/skill tools only when an operator explicitly opts in via
// BOBOT_TOOLS. Trust-boundary decision: keep the safe default unless asked.
//
//   BOBOT_TOOLS unset  -> "--no-tools"      (persona-only, current safe default)
//   BOBOT_TOOLS=a,b    -> "--tools a,b"      (allowlisted extension tools live)
//   BOBOT_TOOLS=all    -> no flag            (every tool enabled; explicit risk)

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

// toolFlags returns the pi flags that control which tools (built-in +
// extension + skill) are available during a turn. See BoBot BOBOT_TOOLS doc above.
func toolFlags() []string {
	switch v := os.Getenv("BOBOT_TOOLS"); v {
	case "":
		return []string{"--no-tools"}
	case "all":
		return nil // no flag -> everything pi would load is enabled
	default:
		return []string{"--tools", v}
	}
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
	args := []string{"-p"}
	args = append(args, toolFlags()...)
	args = append(args,
		"--no-context-files",
		"--session-dir", SessionDir,
		"--session-id", sessionID,
		"--system-prompt", systemPrompt,
		prompt,
	)
	cmd := build(ctx, args...)
	cmd.Stdin = io.NopCloser(bytes.NewReader(nil))

	// stdout = pi's reply text; stderr = diagnostics (e.g. 'Warning: No project
	// session found...'). Keeping them separate stops diagnostics from leaking
	// into the Discord reply; stderr is logged instead. ponytail: no flag suppresses
	// that 'creating a new session' line, so route it to logs, not the user.
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if stderr.Len() > 0 {
		slog.Debug("pi stderr", "session_id", sessionID, "stderr", strings.TrimSpace(stderr.String()))
	}
	if err != nil {
		return stdout.String(), fmt.Errorf("run %q: %w (stderr: %s)",
			strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}