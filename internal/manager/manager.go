package manager

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"initos/internal/config"
)

type Manager interface {
	Install(ctx context.Context, tool config.Tool) error
	IsInstalled(ctx context.Context, tool config.Tool) (bool, error)
	CommandFor(tool config.Tool) (string, error)
	// PlaceholderHomebrewCommand returns the configured darwin install line for macOS (typically brew …).
	// Real Homebrew install orchestration (formula/cask resolution, brew execution) is not implemented yet.
	PlaceholderHomebrewCommand(tool config.Tool) string
}

type MockManager struct {
	goos string
}

func NewManager() Manager {
	return &MockManager{goos: runtime.GOOS}
}

func (m *MockManager) Install(ctx context.Context, tool config.Tool) error {
	command, err := m.CommandFor(tool)
	if err != nil {
		return err
	}
	if command == "" {
		return fmt.Errorf("tool %q has empty install command for %s", tool.ID, m.goos)
	}

	// mock install latency; later replace with real execution + streaming logs.
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(700 * time.Millisecond):
		return nil
	}
}

func (m *MockManager) IsInstalled(ctx context.Context, tool config.Tool) (bool, error) {
	if tool.Check == "" {
		return false, nil
	}
	return runCheckCommand(ctx, tool.Check)
}

func (m *MockManager) CommandFor(tool config.Tool) (string, error) {
	switch m.goos {
	case "darwin":
		return tool.Install.Darwin, nil
	case "linux":
		return tool.Install.Debian, nil
	case "windows":
		return "", nil
	default:
		return "", fmt.Errorf("unsupported OS: %s", m.goos)
	}
}

func (m *MockManager) PlaceholderHomebrewCommand(tool config.Tool) string {
	if m.goos != "darwin" {
		return ""
	}
	cmd, err := m.CommandFor(tool)
	if err != nil || cmd == "" {
		return ""
	}
	return cmd
}

func runCheckCommand(ctx context.Context, command string) (bool, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
