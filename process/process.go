package process

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const LIMIT = 5

var (
	processMutex sync.Mutex // Mutex for managing process states

)

// Process struct to define a process
type Process struct {
	Pid          int         // Process ID
	Name         string      // Process name
	Path         string      // Path to the executable
	Stat         int         // Process status (0: stopped, 1: running)
	Schedul      int         // Scheduling type (0: manual, 1: automatic)
	process      *exec.Cmd   // Executable process
	Timing       *TimingRule // Scheduling rules for the process (optional)
	IsJobDeleted int         // 0 for false and 1 for true
}

// ProcessManager struct to manage multiple processes
type ProcessManager struct {
	Processes []*Process
}

// NewProcess creates a new process with the given parameters
func NewProcess(name, path string, stat, schedul int) *Process {
	return &Process{
		Pid:          0,       // Process ID (initially 0)
		Name:         name,    // Name of the process
		Path:         path,    // Path to the process executable
		Stat:         stat,    // Current status of the process (0: stopped, 1: running)
		Schedul:      schedul, // Scheduling status (0: manual, 1: automatic)
		process:      nil,     // Holds the exec.Cmd instance for the process, initially nil
		Timing:       nil,     // Holds timing rules for scheduled processes, initially nil
		IsJobDeleted: 0,
	}
}

// StartProcess starts the process with specific arguments based on queuing and scheduling
func (p *Process) StartProcess(args ...string) error {
	// Mutex to ensure thread safety
	processMutex.Lock()
	defer processMutex.Unlock()

	// If the process is scheduled, it cannot be started manually
	if p.Schedul == 1 {
		return fmt.Errorf("process %s is scheduled and cannot be started manually", p.Name)
	}

	// If the process is already running, no need to start it again
	if p.Stat == 1 {
		return fmt.Errorf("process %s is already running", p.Name)
	}

	// Use exec.Command to actually start the process
	cmd := exec.Command(p.Path, args...)
	err := cmd.Start() // Start the process
	if err != nil {
		return fmt.Errorf("failed to start process %s: %v", p.Name, err)
	}

	// Save process state to ensure persistence (this could include PID, Stat, etc.)
	p.Pid = cmd.Process.Pid
	p.Stat = 1 // Mark process as running

	// Save the new status
	if err := p.SaveStat(); err != nil {
		return fmt.Errorf("failed to save process status after start: %v", err)
	}

	// Attach process for later management (e.g., waiting for process to finish)
	p.process = cmd

	// Inform the user the process started successfully
	fmt.Printf("Process %s (PID: %d) started successfully.\n", p.Name, p.Pid)

	return nil
}

// Stop stops the process if it is running
func (p *Process) Stop() error {
	// Mutex to ensure thread safety
	processMutex.Lock()
	defer processMutex.Unlock()

	// Check if the process is running
	if p.Stat == 0 {
		return fmt.Errorf("process %s is not running", p.Name)
	}

	// Kill the process
	err := p.process.Process.Kill()
	if err != nil {
		return fmt.Errorf("failed to stop process %s (PID: %d): %v", p.Name, p.Pid, err)
	}

	// Set the status to not running
	p.Stat = 0
	p.Pid = 0

	fmt.Printf("Process %s stopped successfully.\n", p.Name)

	// Save the new status
	if err := p.SaveStat(); err != nil {
		return fmt.Errorf("failed to save process status after stop: %v", err)
	}
	return nil
}

// TimingRule struct to define scheduling rules for a process
type TimingRule struct {
	ScheduleTime time.Time // Exact time the process should run
}

func CreateTimingRule(ruleName string, scheduleInput string) error { //RFC1123 or timestamp
	var scheduleTime time.Time
	var err error

	// Allow users to input either a timestamp or a formatted time string
	if timestamp, err := strconv.ParseInt(scheduleInput, 10, 64); err == nil {
		// Input is a timestamp
		scheduleTime = time.Unix(timestamp, 0)
	} else {
		// Attempt to parse as a formatted string
		scheduleTime, err = time.Parse(time.RFC1123, scheduleInput)
		if err != nil {
			return fmt.Errorf("invalid time format: %w", err)
		}
	}

	rule := TimingRule{
		ScheduleTime: scheduleTime,
	}

	// Construct the file path
	filePath := "./data/time/" + ruleName + ".json"

	// Check if the file already exists
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("timing rule '%s' already exists at %s", ruleName, filePath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check file existence: %w", err)
	}

	// Ensure the directory exists
	dirPath := "./data/time/"
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory '%s': %w", dirPath, err)
	}

	// Save the rule to the file
	errr := SaveToFile(filePath, rule)
	if errr != nil {
		return fmt.Errorf("failed to create timing rule: %w", err)
	}

	// Print success message
	fmt.Printf("Timing rule '%s' created successfully at %s\n", ruleName, scheduleTime.Format(time.RFC1123))
	return nil
}

func (p *Process) SetJob(timingRuleName string) error {
	var rule TimingRule

	// Load timing rule from JSON file
	filePath := "./data/time/" + timingRuleName + ".json"

	if !FileExists(filePath) {
		fmt.Println("TimeRule not found")
		return nil
	}
	err := LoadFromFile(filePath, &rule)
	if err != nil {
		return fmt.Errorf("failed to load timing rule: %w", err)
	}
	if p.Schedul != 1 {
		fmt.Println("This process is mannual")
		return fmt.Errorf("-1")
	}
	// Assign the timing rule to the process
	p.Timing = &rule
	fmt.Printf("Job set for process %s at %s\n", p.Name, p.Timing.ScheduleTime.Format(time.RFC1123))

	savePath := "./schu/" + p.Name + ".json"
	err = SaveToFile(savePath, p)
	if err != nil {
		return fmt.Errorf("failed to save process data: %w", err)
	}

	fmt.Printf("Process %s saved successfully!\n", p.Name)
	return nil
}

// StartJob starts the process based on the timing rule if defined
func (p *Process) StartJob() error {
	if p.IsRunning() {
		fmt.Printf("Process %s is already running\n", p.Name)
		return nil
	}

	if p.Timing != nil {
		scheduleTime := p.Timing.ScheduleTime
		now := time.Now()

		// If the scheduled time is in the future, wait non-blockingly
		if now.Before(scheduleTime) {
			waitDuration := time.Until(scheduleTime)
			fmt.Printf("Waiting %v for process %s to start...\n", waitDuration, p.Name)
			time.AfterFunc(waitDuration, func() {
				if p.IsJobDeleted != 1 {
					if err := p.StartProcess(); err != nil {
						fmt.Printf("Failed to start process %s: %v\n", p.Name, err)
					}
				}
			})
			return nil
		}
	}

	// Start the process immediately if no timing rule or past schedule time
	return p.StartProcess()
}

// ListProcesses lists all processes managed by the ProcessManager
func (pm *ProcessManager) ListProcesses() {
	for _, proc := range pm.Processes {
		fmt.Printf("Process Name: %s, PID: %d, Status: %s, Path: %s\n", proc.Name, proc.Pid, proc.GetStatus(), proc.Path)
	}
}

// Restart restarts the process by stopping and starting it again
func (p *Process) Restart() error {
	if !p.IsRunning() {
		return fmt.Errorf("process %s is not running, start it first", p.Name)
	}
	err := p.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop process %s for restart: %v", p.Name, err)
	}
	return p.StartProcess()
}

func (pm *ProcessManager) AddProcess(name, path string, schedul int) error {
	proc := NewProcess(name, path, 0, schedul)
	err := proc.SaveStat()
	if err != nil {
		fmt.Println("Error in saving status of the process")
		fmt.Printf("%s", err)
	}
	// You can use args later when starting the process, but not when creating it
	pm.Processes = append(pm.Processes, proc)
	return nil
}

// GetStatus returns the current status of the process
func (p *Process) GetStatus() string {
	if p.Stat == 1 {
		return "running"
	}
	return "stopped"
}

// Utility method to check if a process is running based on its PID
func (p *Process) IsRunning() bool {
	_, err := os.FindProcess(p.Pid)
	if err != nil {
		return false // Process is not found, thus not running
	}
	return true
}

// KillProcess terminates a process based on PID
func (pm *ProcessManager) KillProcess(pid int) error {
	for _, p := range pm.Processes {
		if p.Pid == pid {
			if p.Stat == 0 {
				return fmt.Errorf("process %s (PID: %d) is not running", p.Name, p.Pid)
			}

			// Kill the process
			err := p.process.Process.Kill()
			if err != nil {
				return fmt.Errorf("failed to kill process %s (PID: %d): %v", p.Name, p.Pid, err)
			}

			p.Stat = 0 // Update the status to stopped
			return nil
		}
	}
	return fmt.Errorf("process with PID %d not found", pid)
}

func (pm *ProcessManager) RemoveProcess(pid int, name string) error {
	for i, proc := range pm.Processes {
		if (name != "" && proc.Name == name) || (pid != 0 && proc.Pid == pid) {
			// Stop the process if it is running
			if proc.Stat == 1 {
				proc.Stat = 0
				proc.SaveStat()
				err := proc.process.Process.Kill() // Kill the running process
				if err != nil {
					return fmt.Errorf("failed to kill process with PID %d: %v", proc.Pid, err)
				}
			}
			// Remove the process from the slice
			pm.Processes = append(pm.Processes[:i], pm.Processes[i+1:]...)
			fmt.Printf("Process with %s removed successfully.\n", name)
			return nil
		}
	}
	if name != "" {
		return fmt.Errorf("process with name %s not found", name)
	}
	return fmt.Errorf("process with PID %d not found", pid)
}

func (pm *ProcessManager) DeleteJob(pid int, name string) error {
	// Validate inputs
	if pid == 0 && name == "" {
		return fmt.Errorf("both PID and name are empty; provide at least one identifier")
	}

	// Check if there are no processes to manage
	if len(pm.Processes) == 0 {
		return fmt.Errorf("no processes are currently managed")
	}

	// Log the delete action
	fmt.Printf("Attempting to delete job with PID: %d or Name: %s\n", pid, name)

	// Search for the process by PID or name
	for i, proc := range pm.Processes {
		if (pid != 0 && proc.Pid == pid) || (name != "" && proc.Name == name) {
			// Stop the process if it is running
			if proc.IsRunning() {
				if err := proc.Stop(); err != nil {
					return fmt.Errorf("failed to stop running process with PID %d: %v", proc.Pid, err)
				}
			}

			// Remove the process from the slice
			pm.Processes = append(pm.Processes[:i], pm.Processes[i+1:]...)

			// Remove related schedule file, if it exists
			if proc.Schedul == 1 {
				scheduleFilePath := fmt.Sprintf("./schu/%s.json", proc.Name)
				if FileExists(scheduleFilePath) {
					if err := os.Remove(scheduleFilePath); err != nil {
						return fmt.Errorf("failed to delete schedule file for process %s: %v", proc.Name, err)
					}
				}
			}
			//notifying others
			proc.IsJobDeleted = 1

			fmt.Printf("Job for process %s (PID: %d) deleted successfully.\n", proc.Name, proc.Pid)
			return nil
		}
	}

	// If no matching process is found
	if pid != 0 {
		return fmt.Errorf("job with PID %d not found", pid)
	}
	if name != "" {
		return fmt.Errorf("job with name %s not found", name)
	}

	return nil
}

func (pm *ProcessManager) GetProcess(pid int, name string) (*Process, error) {
	// Validate inputs
	if pid == 0 && name == "" {
		return nil, fmt.Errorf("both PID and name are empty; provide at least one identifier")
	}

	// Check if no processes are managed
	if len(pm.Processes) == 0 {
		return nil, fmt.Errorf("no processes are currently managed")
	}

	// Log the search action
	fmt.Printf("Searching for process with PID: %d or Name: %s\n", pid, name)

	// Search by PID if name is "no"
	if name == "no" {
		for _, proc := range pm.Processes {
			if proc.Pid == pid {
				return proc, nil // Process found
			}
		}
		return nil, fmt.Errorf("process with PID %d not found", pid) // Process not found
	}

	// Otherwise, search by name
	for _, proc := range pm.Processes {
		if proc.Name == name {
			return proc, nil // Process found
		}
	}

	// Process not found
	return nil, fmt.Errorf("process with name %s or PID %d not found", name, pid)
}

// ShowSystemInfo displays system information (for now, just an example)
func (pm *ProcessManager) ShowSystemInfo() {
	fmt.Println("System Information:")

	// Get and display number of CPUs
	numCPU := runtime.NumCPU()
	fmt.Printf("Number of CPUs: %d\n", numCPU)

	// Get and display system time
	systemTime := time.Now().Format(time.RFC1123)
	fmt.Printf("Current System Time: %s\n", systemTime)

	// Get and display memory statistics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	fmt.Printf("Memory Usage: Alloc = %d KB, TotalAlloc = %d KB, Sys = %d KB, NumGC = %d\n",
		memStats.Alloc/1024, memStats.TotalAlloc/1024, memStats.Sys/1024, memStats.NumGC)

	// Display Go version
	fmt.Printf("Go Version: %s\n", runtime.Version())
}
