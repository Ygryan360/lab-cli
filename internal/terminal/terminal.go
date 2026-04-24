package terminal

import (
	"fmt"
	"os"
	"os/exec"
)

// OpenTab opens a new terminal tab in the given directory.
// It detects the terminal from config, falling back to env detection.
func OpenTab(termName, projectPath, projectName string) error {
	switch termName {
	case "kitty", "":
		return openKitty(projectPath, projectName)
	case "gnome-terminal":
		return openGnomeTerminal(projectPath)
	case "wezterm":
		return openWezterm(projectPath, projectName)
	default:
		// Generic fallback: try kitty, then xterm
		return openKitty(projectPath, projectName)
	}
}

// openKitty opens a new tab in the current kitty window via remote control.
// Requires: allow_remote_control yes in kitty.conf
func openKitty(path, title string) error {
	if _, err := exec.LookPath("kitty"); err != nil {
		return fmt.Errorf("kitty not found in PATH")
	}

	args := []string{
		"@", "launch",
		"--type=tab",
		"--tab-title", title,
		"--cwd", path,
	}

	cmd := exec.Command("kitty", args...)
	cmd.Env = os.Environ()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("kitty @ launch failed: %w", err)
	}
	return nil
}

func openGnomeTerminal(path string) error {
	cmd := exec.Command("gnome-terminal", "--working-directory="+path)
	return cmd.Start()
}

func openWezterm(path, title string) error {
	cmd := exec.Command("wezterm", "cli", "spawn",
		"--new-window=false",
		"--cwd", path)
	return cmd.Start()
}

// Detect tries to identify the running terminal from environment variables
func Detect() string {
	if t := os.Getenv("TERM_PROGRAM"); t != "" {
		return t
	}
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return "kitty"
	}
	if os.Getenv("WEZTERM_PANE") != "" {
		return "wezterm"
	}
	return "kitty" // default
}