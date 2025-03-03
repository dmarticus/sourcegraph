// Package usershell gathers information on the current user and injects then in a context.Context.
package usershell

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

type key struct{}

// userShell stores which shell and which configuration file a user is using.
type userShell struct {
	shellPath       string
	shellConfigPath string
}

// ShellPath returns the path to the shell used by the current unix user.
func ShellPath(ctx context.Context) string {
	v := ctx.Value(key{}).(userShell)
	return v.shellPath
}

// ShellPath returns the path to the shell configuration (bashrc...) used by the current unix user.
func ShellConfigPath(ctx context.Context) string {
	v := ctx.Value(key{}).(userShell)
	return v.shellConfigPath
}

// GuessUserShell inspect the current environment to infer the shell the current user is running
// and which configuration file it depends on.
func GuessUserShell() (string, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	// Look up which shell the user is using, because that's most likely the
	// one that has all the environment correctly setup.
	shell, ok := os.LookupEnv("SHELL")
	var shellrc string
	if !ok {
		// If we can't find the shell in the environment, we fall back to `bash`
		shell = "bash"
	}
	switch {
	case strings.Contains(shell, "bash"):
		shellrc = ".bashrc"
	case strings.Contains(shell, "zsh"):
		shellrc = ".zshrc"
	}
	return shell, filepath.Join(home, shellrc), nil
}

// Context extends ctx with the UserContext of the current user.
func Context(ctx context.Context) (context.Context, error) {
	shell, shellConfigPath, err := GuessUserShell()
	if err != nil {
		return nil, err
	}
	userCtx := userShell{
		shellPath:       shell,
		shellConfigPath: shellConfigPath,
	}
	return context.WithValue(ctx, key{}, userCtx), nil
}
