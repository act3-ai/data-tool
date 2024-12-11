package secret

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	"git.act3-ace.com/ace/go-common/pkg/redact"
)

// PromptUsername prompts a username input from stdin.
func PromptUsername(ctx context.Context, out io.Writer) (string, error) {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "prompting username for registry auth")
	_, err := fmt.Fprint(out, "Username: ")
	if err != nil {
		return "", err
	}
	reader := bufio.NewReader(os.Stdin)
	line, _, err := reader.ReadLine()
	if err != nil {
		return "", fmt.Errorf("error reading from stdin: %w", err)
	}
	username := strings.TrimSpace(string(line))

	return username, nil
}

// PromptPassword prompts a password input from stdin.
func PromptPassword(ctx context.Context, out io.Writer) (redact.Secret, error) {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "prompting password for registry auth")
	_, err := fmt.Fprint(out, "Password: ")
	if err != nil {
		return "", err
	}
	bpw, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("reading password from term: %w", err)
	}
	password := redact.Secret(bpw)
	if password == "" {
		return "", fmt.Errorf("password is required")
	}

	return password, nil
}
