package command

import (
	"ExeProcessManager/process"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var manager = process.ProcessManager{}

// Maps the status integer to a readable state (e.g., "running", "stopped")
func mapStat(stat int) string {
	if stat == 1 {
		return "running"
	}
	return "stopped"
}

// StartCLI initializes the command-line interface
func StartCLI() {
	for {
		fmt.Print(">>> ")
		reader := bufio.NewReader(os.Stdin)
		cmd, _ := reader.ReadString('\n')     // خواندن ورودی چندکلمه‌ای
		handleCommand(strings.TrimSpace(cmd)) // حذف فضای اضافی
	}
}

// handleCommand processes user input
func handleCommand(cmd string) {

	args := strings.Fields(cmd)

	if len(args) == 0 {
		return
	}

	switch args[0] {
	case "help":
		showHelp()

	case "list":
		manager.ListProcesses()

	case "stop":
		if len(args) < 2 {
			fmt.Println("Usage: stop <PID>")
			return
		}
		if pid, err := strconv.Atoi(args[1]); err == nil {
			err := manager.KillProcess(pid)
			if err != nil {
				fmt.Println("Error stopping process:", err)
			} else {
				fmt.Println("Process stopped successfully.")
			}
		} else {
			fmt.Println("Invalid PID format.")
		}

	case "start":
		if len(args) < 2 {
			fmt.Println("Usage: start <process_name>")
			return
		}

		// Start process with given name
		name := args[1] // Use the name of the process directly
		proc, err := manager.GetProcess(0, name)
		if err != nil {
			fmt.Println("Error retrieving process:", err)
			return
		}
		err = proc.StartProcess() // Assuming StartProcess method exists
		if err != nil {
			fmt.Println("Error starting process:", err)
		} else {
			fmt.Println("Process started successfully.")
		}

	case "add":
		if len(args) < 4 {
			fmt.Println("Usage: add <name> <path> <schedul>")
			return
		}

		// Print received arguments for debugging
		fmt.Println("Arguments received:", args)

		// Convert schedul to int
		schedul, err := strconv.Atoi(args[3])
		if err != nil {
			fmt.Println("Invalid value for schedul, expected an integer.")
			return
		}

		// Add the process without starting it
		err = manager.AddProcess(args[1], args[2], schedul)
		if err != nil {
			fmt.Println("Error adding process:", err)
		} else {
			fmt.Println("Process added successfully. To start it, use 'start <PID>'.")
		}

	case "sysinfo":
		manager.ShowSystemInfo()

	case "exit":
		fmt.Println("Exiting the program...")
		exitProgram()

	case "settimerule": // name, timeInput
		if len(args) < 3 {
			fmt.Println("Usage: settimerule <name> <timeInput>")
			return
		}
		ruleName := args[1]
		timeInput := args[2]

		// Call CreateTimingRule to create the timing rule
		err := process.CreateTimingRule(ruleName, timeInput)
		if err != nil {
			fmt.Println("Error creating timing rule:", err)
			return
		}

		fmt.Printf("TimingRule %s created successfully!\n", ruleName)

	case "setjobtime": // process name, timing rule
		if len(args) < 3 {
			fmt.Println("Usage: setJobTiming <process_name> <timingRule>")
			return
		}

		name := args[1]
		timingRule := args[2]

		// Get the process by name
		proc, err := manager.GetProcess(0, name)
		if err != nil {
			fmt.Println("Error searching for the process.")
			return
		}

		// Set the job using the provided timing rule
		err = proc.SetJob(timingRule)
		if err != nil {
			fmt.Println("Failed to set job:", err)
			return
		}
		fmt.Printf("Timing job for process %s set successfully!\n", name)

	case "startjob":
		if len(args) < 2 {
			fmt.Println("Usage: startjob <process_name>")
			return
		}

		name := args[1]

		// Get the process by name
		proc, err := manager.GetProcess(0, name)
		if err != nil {
			fmt.Printf("Error retrieving process: %s\n", err)
			return
		}

		// Check if the process is schedulable (Schedul == 1 means it's allowed to be started)
		if proc.Schedul != 1 {
			fmt.Println("The process is not schedulable or is already in an invalid state.")
			return
		}

		// Check if the process is already running
		if proc.Stat == 1 {
			fmt.Println("The process is already running.")
			return
		}

		// If it's not running, attempt to start the process
		err = proc.StartJob() // Assuming StartJob initiates the process
		if err != nil {
			fmt.Printf("Error starting process %s: %s\n", name, err)
			return
		}

		// Confirm that the process has started
		fmt.Printf("Process %s started successfully.\n", name)

	case "status":
		if len(args) < 2 {
			fmt.Println("Usage: status <PID or name>")
			return
		}

		// Try to get process by PID if the argument is an integer (PID)
		if pid, err := strconv.Atoi(args[1]); err == nil {
			proc, err := manager.GetProcess(pid, "no") // Get process by PID
			if err != nil {
				fmt.Println("Error retrieving process:", err)
				return
			}
			fmt.Printf("Process Status for PID %d:\n", pid)
			fmt.Printf("Name: %s\n", proc.Name)
			fmt.Printf("State: %s\n", mapStat(proc.Stat))
			fmt.Printf("Schedulable: %v\n", proc.Schedul == 1)
		} else {
			// Otherwise, try to get process by name
			name := args[1]
			proc, err := manager.GetProcess(0, name) // Get process by name
			if err != nil {
				fmt.Println("Error retrieving process:", err)
				return
			}
			fmt.Printf("Process Status for %s:\n", name)
			fmt.Printf("PID: %d\n", proc.Pid)
			fmt.Printf("State: %s\n", mapStat(proc.Stat))
			fmt.Printf("Schedulable: %v\n", proc.Schedul == 1)
		}

	case "remove":
		if len(args) < 2 {
			fmt.Println("Usage: remove <PID or Name>")
			return
		}

		if pid, err := strconv.Atoi(args[1]); err == nil {
			// If it's a valid integer, treat it as a PID
			err = manager.RemoveProcess(pid, "")
			if err != nil {
				fmt.Println("Error removing process:", err)
			}
			fmt.Println("removed succesfully!")
		} else {
			// Otherwise, treat it as a Name
			err := manager.RemoveProcess(0, args[1])
			if err != nil {
				fmt.Println("Error removing process:", err)
			}
			fmt.Println("removed succesfully!")
		}

	default:
		fmt.Println("Unknown command. Use 'help' for a list of commands.")
	}
}

// showHelp displays available commands
func showHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  help                 - Show available commands")
	fmt.Println("  list                 - List all running processes")
	fmt.Println("  stop <PID>           - Terminate a process by PID")
	fmt.Println("  add <name> <path> <schedul> <requireQueue> - Add a new process (without starting it)")
	fmt.Println("  start <PID>          - Start a process with the given PID")
	fmt.Println("  sysinfo              - Display system information")
	fmt.Println("  exit                 - Exit the application")
	fmt.Println("  settimerule <name> <timeInput>  - Set a timing rule for a process / timeInput : either timestamp or rfc1123 ")
	fmt.Println("  setjobtime <process_name> <timingRule>  - Set a job time rule for a process")
	fmt.Println("  startjob <process_name>  - Start a scheduled job for the process")
}

func exitProgram() {
	exitCode := 0
	fmt.Println("Goodbye!")
	systemExit(exitCode)
}

// systemExit is a placeholder for the real os.Exit
var systemExit = func(code int) {
	// Use os.Exit in actual implementation
}
