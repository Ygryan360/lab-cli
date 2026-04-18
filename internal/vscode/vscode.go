package vscode

import (
	"os/exec"
)

// Open launches VSCode for a given path with an optional profile
func Open(path, profile string) error {
	args := []string{path}
	if profile != "" && profile != "Default" {
		args = append([]string{"--profile", profile}, args...)
	}
	cmd := exec.Command("code", args...)
	return cmd.Start()
}

// IsInstalled checks if the `code` binary is available
func IsInstalled() bool {
	_, err := exec.LookPath("code")
	return err == nil
}
