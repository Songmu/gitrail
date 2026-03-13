package gitrail

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func gitEnv() []string {
	base := os.Environ()
	env := make([]string, 0, len(base)+2)
	env = append(env, "LANG=C", "LC_ALL=C")
	for _, e := range base {
		if strings.HasPrefix(e, "LANG=") || strings.HasPrefix(e, "LC_ALL=") {
			continue
		}
		env = append(env, e)
	}
	return env
}

func gitCmd(ctx context.Context, dir string, errStream io.Writer, args ...string) (string, error) {
	cmdArgs := args
	if dir != "" {
		cmdArgs = append([]string{"-C", dir}, args...)
	}
	cmd := exec.CommandContext(ctx, "git", cmdArgs...)
	cmd.Env = gitEnv()
	cmd.Stderr = errStream
	var buf bytes.Buffer
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return "", exitErrorf(1, "git %s failed: %v", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(buf.String()), nil
}

func gitCmdAllowFail(ctx context.Context, dir string, errStream io.Writer, args ...string) (string, int, error) {
	cmdArgs := args
	if dir != "" {
		cmdArgs = append([]string{"-C", dir}, args...)
	}
	cmd := exec.CommandContext(ctx, "git", cmdArgs...)
	cmd.Env = gitEnv()
	cmd.Stderr = errStream
	var buf bytes.Buffer
	cmd.Stdout = &buf
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return "", -1, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
		}
	}
	return strings.TrimSpace(buf.String()), exitCode, nil
}
