package command

import (
	"ExeProcessManager/config"
	"ExeProcessManager/process"
	"bytes"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
)

// setupCLITest is a helper to create all components for a CLI test.
func setupCLITest(t *testing.T) (*CLI, *process.ProcessManager) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil)) // Discard logs
	cfg := &config.Config{
		DataDir:     t.TempDir(),
		ScheduleDir: t.TempDir(),
	}
	pm := process.NewProcessManager(logger, cfg)
	cli := NewCLI(pm, logger)
	return cli, pm
}

// captureOutput executes a function and captures everything written to Stdout.
func captureOutput(f func()) string {
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f() // Execute the function that prints to stdout

	w.Close()
	os.Stdout = originalStdout // Restore original stdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// TestCLI_AddCommand tests the 'add' command.
func TestCLI_AddCommand(t *testing.T) {
	cli, manager := setupCLITest(t)

	// Capture the output of the handleCommand function
	output := captureOutput(func() {
		cli.handleCommand("add cli-proc /usr/bin/top 0")
	})

	// Assert that the success message was printed
	if !strings.Contains(output, "Process 'cli-proc' added successfully") {
		t.Errorf("expected success message not found in output: got '%s'", output)
	}

	// Assert that the process was actually added to the manager
	_, err := manager.GetProcessByName("cli-proc")
	if err != nil {
		t.Errorf("process was not added to manager after 'add' command: %v", err)
	}
}

// TestCLI_ListCommand tests the 'list' command.
func TestCLI_ListCommand(t *testing.T) {
	cli, manager := setupCLITest(t)

	// Test with no processes
	outputEmpty := captureOutput(func() {
		cli.handleCommand("list")
	})
	if !strings.Contains(outputEmpty, "No processes are being managed") {
		t.Errorf("expected empty message not found: got '%s'", outputEmpty)
	}

	// Add a process and test again
	_, _ = manager.AddProcess("listed-proc", "/bin/ls", 0)
	outputWithProcess := captureOutput(func() {
		cli.handleCommand("list")
	})

	if !strings.Contains(outputWithProcess, "listed-proc") {
		t.Errorf("expected process name not found in list output: got '%s'", outputWithProcess)
	}
}