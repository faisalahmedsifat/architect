package commands

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

func WatchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "watch",
		Short: "Watch for specification changes",
		Long:  "Auto-syncs rules when specifications change",
		RunE:  runWatch,
	}
}

func runWatch(cmd *cobra.Command, args []string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	// Watch .architect directory
	architectDir := ".architect"
	if err := watcher.Add(architectDir); err != nil {
		return fmt.Errorf("failed to watch directory: %w", err)
	}

	color.Yellow("👀 Watching .architect/ for changes...")
	fmt.Println("Press Ctrl+C to stop watching\n")

	// Debounce timer to avoid multiple syncs
	var debounceTimer *time.Timer
	syncFunc := func() {
		if debounceTimer != nil {
			debounceTimer.Stop()
		}
		debounceTimer = time.AfterFunc(500*time.Millisecond, func() {
			timestamp := time.Now().Format("15:04:05")
			color.Blue("[%s] Syncing specifications...", timestamp)
			if err := runSync(cmd, args); err != nil {
				color.Red("Error syncing: %v", err)
			}
		})
	}

	// Watch for events
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				// Only sync for .md and .yaml files
				ext := filepath.Ext(event.Name)
				if ext == ".md" || ext == ".yaml" || ext == ".yml" {
					timestamp := time.Now().Format("15:04:05")
					color.Cyan("[%s] Changed: %s", timestamp, filepath.Base(event.Name))
					syncFunc()
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("Watch error: %v", err)
		}
	}
}
