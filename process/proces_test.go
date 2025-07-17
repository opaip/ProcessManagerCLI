package process

import (
	"ExeProcessManager/config"
	"log/slog"
	"os"
	"testing"
	"time"
)

// setupTestManager is a helper to create a manager for testing.
func setupTestManager(t *testing.T) *ProcessManager {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := &config.Config{
		DataDir:     t.TempDir(), // Use a temporary directory for test artifacts
		ScheduleDir: t.TempDir(),
	}
	return NewProcessManager(logger, cfg)
}

func TestNewProcess(t *testing.T) {
	pm := setupTestManager(t)
	p := pm.NewProcess("test", "/bin/true", 0)

	if p.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", p.Name)
	}
	if p.Stat != 0 {
		t.Errorf("expected initial status 0, got %d", p.Stat)
	}
	if p.manager == nil {
		t.Error("process manager reference was not set")
	}
}

func TestProcessStartStop(t *testing.T) {
	// This test is OS-dependent. 'sleep' is common on Unix-like systems.
	pm := setupTestManager(t)
	p, err := pm.AddProcess("sleeper", "sleep", 0)
	if err != nil {
		t.Fatalf("failed to add process: %v", err)
	}

	// Test Start
	err = p.Start("10") // Start 'sleep 10'
	if err != nil {
		t.Fatalf("failed to start process: %v", err)
	}

	if p.Stat != 1 {
		t.Error("process status should be 1 (running) after start")
	}
	if p.Pid == 0 {
		t.Error("process PID should not be 0 after start")
	}

	// Give the OS a moment to register the process
	time.Sleep(100 * time.Millisecond)

	// Test Stop
	err = p.Stop()
	if err != nil {
		t.Fatalf("failed to stop process: %v", err)
	}

	if p.Stat != 0 {
		t.Error("process status should be 0 (stopped) after stop")
	}
	if p.Pid != 0 {
		t.Error("process PID should be 0 after stop")
	}
}

func TestAddProcessDuplicate(t *testing.T) {
	pm := setupTestManager(t)
	_, err := pm.AddProcess("duplicate", "/bin/true", 0)
	if err != nil {
		t.Fatalf("first add should succeed: %v", err)
	}

	_, err = pm.AddProcess("duplicate", "/bin/true", 0)
	if err == nil {
		t.Fatal("second add with same name should fail, but it succeeded")
	}
}