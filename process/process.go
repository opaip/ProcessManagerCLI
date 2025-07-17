package process

import (
	"ExeProcessManager/config"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// Process struct defines a manageable process.
type Process struct {
	Pid     int    `json:"pid"`
	Name    string `json:"name"`
	Path    string `json:"path"`
	Stat    int    `json:"stat"`    // 0: stopped, 1: running
	Schedul int    `json:"schedul"` // 0: manual, 1: automatic

	// Non-exported fields
	process      *exec.Cmd   `json:"-"` // The running command
	Timing       *TimingRule `json:"timing,omitempty"`
	IsJobDeleted int         `json:"is_job_deleted"`
	manager      *ProcessManager `json:"-"` // Reference to the manager for config/logging
}

// ProcessManager manages all processes.
type ProcessManager struct {
	Processes    []*Process
	logger       *slog.Logger
	config       *config.Config
	processMutex sync.Mutex
}

// NewProcessManager creates a new instance of ProcessManager.
func NewProcessManager(logger *slog.Logger, cfg *config.Config) *ProcessManager {
	return &ProcessManager{
		Processes: make([]*Process, 0),
		logger:    logger,
		config:    cfg,
	}
}

// NewProcess creates a new process instance.
func (pm *ProcessManager) NewProcess(name, path string, schedul int) *Process {
	return &Process{
		Pid:     0,
		Name:    name,
		Path:    path,
		Stat:    0,
		Schedul: schedul,
		manager: pm, // Link back to the manager
	}
}

// AddProcess creates a new process and adds it to the manager.
func (pm *ProcessManager) AddProcess(name, path string, schedul int) (*Process, error) {
	pm.processMutex.Lock()
	defer pm.processMutex.Unlock()

	// Check for existing process with the same name
	for _, p := range pm.Processes {
		if p.Name == name {
			return nil, fmt.Errorf("process with name '%s' already exists", name)
		}
	}

	proc := pm.NewProcess(name, path, schedul)
	if err := proc.SaveState(); err != nil {
		return nil, fmt.Errorf("failed to save initial process state: %w", err)
	}

	pm.Processes = append(pm.Processes, proc)
	pm.logger.Info("process added successfully", "name", name, "path", path)
	return proc, nil
}

// Start starts a manually-controlled process.
func (p *Process) Start(args ...string) error {
	p.manager.processMutex.Lock()
	defer p.manager.processMutex.Unlock()

	if p.Schedul == 1 {
		return fmt.Errorf("process '%s' is scheduled and cannot be started manually", p.Name)
	}
	if p.Stat == 1 {
		return fmt.Errorf("process '%s' is already running with PID %d", p.Name, p.Pid)
	}

	cmd := exec.Command(p.Path, args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process executable: %w", err)
	}

	p.process = cmd
	p.Pid = cmd.Process.Pid
	p.Stat = 1 // Mark as running

	p.manager.logger.Info("process started successfully", "name", p.Name, "pid", p.Pid)

	// Persist the new state
	return p.SaveState()
}

// Stop terminates the process.
func (p *Process) Stop() error {
	p.manager.processMutex.Lock()
	defer p.manager.processMutex.Unlock()

	if p.Stat == 0 {
		return fmt.Errorf("process '%s' is not running", p.Name)
	}

	// On Windows, FindProcess is needed. On Unix, the process handle is sufficient.
	osProc, err := os.FindProcess(p.Pid)
	if err != nil {
		// If the process doesn't exist, it might have already terminated.
		p.manager.logger.Warn("could not find process to stop, it may have already exited", "name", p.Name, "pid", p.Pid)
		p.Stat = 0
		p.Pid = 0
		return p.SaveState()
	}

	if err := osProc.Kill(); err != nil {
		return fmt.Errorf("failed to kill process: %w", err)
	}

	// Wait for the process to release resources, optional but good practice
	_, _ = p.process.Process.Wait()

	p.manager.logger.Info("process stopped successfully", "name", p.Name, "pid", p.Pid)

	p.Stat = 0
	p.Pid = 0
	p.process = nil

	return p.SaveState()
}

// GetStatus returns a human-readable status string.
func (p *Process) GetStatus() string {
	if p.Stat == 1 {
		// We can also ping the process to be sure
		proc, err := os.FindProcess(p.Pid)
		if err != nil || proc == nil {
			return "stopped (crashed)"
		}
		return "running"
	}
	return "stopped"
}

// RemoveProcess finds a process by name and removes it from the manager.
func (pm *ProcessManager) RemoveProcess(name string) error {
	pm.processMutex.Lock()
	defer pm.processMutex.Unlock()

	for i, p := range pm.Processes {
		if p.Name == name {
			if p.Stat == 1 {
				// Use the official Stop method to ensure clean state change
				if err := p.Stop(); err != nil {
					pm.logger.Error("failed to stop process during removal, attempting to continue", "name", p.Name, "error", err)
				}
			}

			// Remove process state file
			if err := p.DeleteStateFile(); err != nil {
				pm.logger.Error("failed to delete process state file", "name", p.Name, "error", err)
				// Continue with removal from memory regardless
			}
			
			// Remove schedule file if it exists
			if p.Schedul == 1 {
				scheduleFilePath := filepath.Join(pm.config.ScheduleDir, p.Name+".json")
				if FileExists(scheduleFilePath) {
					if err := os.Remove(scheduleFilePath); err != nil {
						pm.logger.Error("failed to delete schedule file", "name", p.Name, "path", scheduleFilePath, "error", err)
					}
				}
			}


			// Remove from the slice
			pm.Processes = append(pm.Processes[:i], pm.Processes[i+1:]...)
			pm.logger.Info("process removed successfully", "name", name)
			return nil
		}
	}
	return fmt.Errorf("process with name '%s' not found", name)
}

// GetProcessByName finds and returns a process by its name.
func (pm *ProcessManager) GetProcessByName(name string) (*Process, error) {
	pm.processMutex.Lock()
	defer pm.processMutex.Unlock()

	for _, p := range pm.Processes {
		if p.Name == name {
			return p, nil
		}
	}
	return nil, fmt.Errorf("process with name '%s' not found", name)
}

// SaveState saves the process's current state to a JSON file.
func (p *Process) SaveState() error {
	dataDir := p.manager.config.DataDir
	stateFilePath := filepath.Join(dataDir, "processes", p.Name+".json")

	if err := os.MkdirAll(filepath.Dir(stateFilePath), 0755); err != nil {
		return fmt.Errorf("failed to create process state directory: %w", err)
	}

	return SaveToFile(stateFilePath, p)
}

// DeleteStateFile removes the process's state file.
func (p *Process) DeleteStateFile() error {
    dataDir := p.manager.config.DataDir
	stateFilePath := filepath.Join(dataDir, "processes", p.Name+".json")
	if !FileExists(stateFilePath) {
        return nil // Nothing to delete
    }
	return os.Remove(stateFilePath)
}


// LoadProcessesFromDisk scans the process data directory and loads all processes into the manager.
func (pm *ProcessManager) LoadProcessesFromDisk() error {
	pm.processMutex.Lock()
	defer pm.processMutex.Unlock()

	processDir := filepath.Join(pm.config.DataDir, "processes")
	if !FileExists(processDir) {
		pm.logger.Info("process data directory does not exist, skipping load", "path", processDir)
		return nil
	}

	files, err := os.ReadDir(processDir)
	if err != nil {
		return fmt.Errorf("failed to read process directory: %w", err)
	}

	loadedCount := 0
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(processDir, file.Name())
		proc := &Process{manager: pm} // Create new process with manager reference

		if err := LoadFromFile(filePath, proc); err != nil {
			pm.logger.Warn("failed to load process state from file, skipping", "file", filePath, "error", err)
			continue
		}
		
		// Reset state on load - assume all processes are stopped initially.
		// A more advanced system could check if the PID is still active.
		proc.Stat = 0
		proc.Pid = 0
		proc.process = nil

		pm.Processes = append(pm.Processes, proc)
		loadedCount++
	}

	if loadedCount > 0 {
		pm.logger.Info("loaded processes from disk", "count", loadedCount)
	}
	return nil
}