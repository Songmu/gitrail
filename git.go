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

func gitCmd(ctx context.Context, dir string, errStream io.Writer, args ...string) (string, error) {
	cmdArgs := args
	if dir != "" {
		cmdArgs = append([]string{"-C", dir}, args...)
	}
	cmd := exec.CommandContext(ctx, "git", cmdArgs...)
	cmd.Env = append(os.Environ(), "LANG=C")
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
	cmd.Env = append(os.Environ(), "LANG=C")
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
