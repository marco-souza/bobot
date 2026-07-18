// Package pi spawns pi coding-agent instances to answer user messages.
package pi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

// systemPrompt is the assistant persona pi adopts for Discord replies.
const systemPrompt = "You are bobot, a concise assistant in a Discord chat. " +
	"Reply briefly and conversationally. No code blocks unless asked."

// Ask runs a non-interactive pi instance with prompt and returns its output.
// It is the one seam between bobot and the pi binary: everything upstream just
// hands it a prompt string, everything downstream consumes a reply string.
func Ask(ctx context.Context, prompt string) (string, error) {
	out, err := run(ctx, buildCmd, prompt)
	if err != nil {
		return "", fmt.Errorf("pi: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// AskTimeout is Ask with a bounded deadline.
func AskTimeout(prompt string, d time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	return Ask(ctx, prompt)
}

// commandFunc builds the exec.Cmd for a prompt; overridable in tests.
type commandFunc func(ctx context.Context, args ...string) *exec.Cmd

var buildCmd commandFunc = func(ctx context.Context, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, "pi", args...)
}

// run executes buildCmd for prompt and returns its stdout.
func run(ctx context.Context, build commandFunc, prompt string) (string, error) {
	// ponytail: no-tools/no-context-files/no-session keeps this a pure chat turn,
	// no file access, no project AGENTS.md leaking, no clutter on disk. Upgrade
	// to a persistent per-thread session if conversation continuity is wanted.
	args := []string{
		"-p",
		"--no-tools",
		"--no-context-files",
		"--no-session",
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