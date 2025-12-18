package ui

import (
	"os"
	"os/exec"
)

func openEditor() (string, error) {
	tempFile, err := os.CreateTemp("", "mynote-*.md")
	if err != nil {
		return "", err
	}
	defer os.Remove(tempFile.Name())

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim" // Fallback
	}

	cmd := exec.Command(editor, tempFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(tempFile.Name())
	return string(content), err
}
