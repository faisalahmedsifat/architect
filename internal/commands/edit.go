package commands

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

func EditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Edit specifications",
		Long:  "Opens your specifications in the default editor",
		RunE:  runEdit,
	}
}

func runEdit(cmd *cobra.Command, args []string) error {
	var choice string
	prompt := &survey.Select{
		Message: "What would you like to edit?",
		Options: []string{
			"Project description (.architect/project.md)",
			"API specifications (.architect/api.yaml)",
			"Both files",
		},
	}
	survey.AskOne(prompt, &choice)

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("ARCHITECT_EDITOR")
	}
	if editor == "" {
		// Try common editors
		editors := []string{"code", "vim", "nano", "notepad"}
		for _, e := range editors {
			if _, err := exec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
	}

	if editor == "" {
		return fmt.Errorf("no editor found. Set EDITOR or ARCHITECT_EDITOR environment variable")
	}

	var files []string
	switch choice {
	case "Project description (.architect/project.md)":
		files = []string{".architect/project.md"}
	case "API specifications (.architect/api.yaml)":
		files = []string{".architect/api.yaml"}
	case "Both files":
		files = []string{".architect/project.md", ".architect/api.yaml"}
	}

	for _, file := range files {
		cmd := exec.Command(editor, file)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to open editor: %w", err)
		}
	}

	// Ask if they want to sync after editing
	var doSync bool
	syncPrompt := &survey.Confirm{
		Message: "Would you like to sync Cursor rules now?",
		Default: true,
	}
	survey.AskOne(syncPrompt, &doSync)

	if doSync {
		return runSync(cmd, args)
	}

	return nil
}
