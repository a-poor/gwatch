package main

import (
	"context"
	"flag"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var (
	interval time.Duration
)

func init() {
	// Set a custom usage function
	flag.Usage = CustomUsage

	// Register flags
	flag.DurationVar(&interval, "i", time.Second, "How often should the command be run?")
}

func main() {
	// Parse the flags
	flag.Parse()

	// Get the command to run
	args := flag.Args()

	// Check that a command was
	if len(args) == 0 {
		fmt.Println("Error: No command specified")
		os.Exit(0)
	}

	// Get a base context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the model to view the output
	m := newModel(
		strings.Join(args, " "),
		cancel,
		interval,
	)

	// Create the bubbletea program from model
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Create a ticker, using the specified interval, for running
	// the command in the background
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var wg sync.WaitGroup

	// Start running the command in the background
	var bgErr error
	go func() {
		wg.Add(1)
		defer wg.Done()
		defer ticker.Stop()
		defer fmt.Println("Inner complete!")

		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				// Create the command
				cmd := exec.CommandContext(ctx, args[0], args[1:]...)

				// Run the command and capture the output
				out, err := cmd.CombinedOutput()
				if err != nil {
					// Trigger context cancel
					cancel()

					// Trigger bubbletea app quit
					p.Quit()

					// Set the error to be used after done
					bgErr = err
					return
				}

				fmt.Printf("Time: %s\nOut:%q\n", time.Now(), string(out))

				// Set the output
				m.updateOutput(string(out))
			}
		}
	}()

	// Run and wait the handle errors (from the TUI or bg executor)
	//err := p.Start()
	//if err != nil {
	//	fmt.Printf("Error with TUI: %s\n", err)
	//}
	time.Sleep(time.Second * 5)
	if bgErr != nil {
		fmt.Printf("Error executing cmd: %s\n", bgErr)
	}
	cancel()
	wg.Wait()
}
