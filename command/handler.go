package command

import (
	"ExeProcessManager/process"
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// CLI handles the command-line interface.
type CLI struct {
	manager *process.ProcessManager
	logger  *slog.Logger
}

// NewCLI creates a new CLI handler.
func NewCLI(manager *process.ProcessManager, logger *slog.Logger) *CLI {
	return &CLI{
		manager: manager,
		logger:  logger,
	}
}

// Start begins the CLI read-eval-print loop (REPL).
func (cli *CLI) Start(ctx context.Context) {
	cli.logger.Info("CLI started. Type 'help' for commands.")
	reader := bufio.NewReader(os.Stdin)

	for {
		select {
		case <-ctx.Done(): // Check if a shutdown has been requested
			cli.logger.Info("CLI shutting down.")
			return
		default:
			fmt.Print(">>> ")
			cmd, err := reader.ReadString('\n')
			if err != nil {
				// This can happen if Stdin is closed, e.g., during shutdown
				cli.logger.Debug("CLI reader error", "error", err)
				return
			}
			cli.handleCommand(strings.TrimSpace(cmd))
		}
	}
}

func (cli *CLI) handleCommand(cmd string) {
	args := strings.Fields(cmd)
	if len(args) == 0 {
		return
	}

	command := args[0]
	params := args[1:]

	switch command {
	case "help":
		showHelp()
	case "exit":
		fmt.Println("Please use Ctrl+C to exit the application gracefully.")
	case "add":
		cli.addProcess(params)
	case "start":
		cli.startProcess(params)
	case "stop":
		cli.stopProcess(params)
	case "status":
		cli.showStatus(params)
	case "list":
		cli.listProcesses()
	case "remove":
		cli.removeProcess(params)
	// Add other cases for scheduling here...
	default:
		fmt.Println("Unknown command. Use 'help' for a list of commands.")
	}
}

// --- Command Implementations ---

func (cli *CLI) addProcess(params []string) {
	if len(params) < 3 {
		fmt.Println("Usage: add <name> <path> <schedul (0=manual, 1=auto)>")
		return
	}
	name, path := params[0], params[1]
	schedul, err := strconv.Atoi(params[2])
	if err != nil || (schedul != 0 && schedul != 1) {
		fmt.Println("Invalid value for schedul, must be 0 or 1.")
		return
	}

	if _, err := cli.manager.AddProcess(name, path, schedul); err != nil {
		cli.logger.Error("failed to add process", "error", err)
		fmt.Println("Error:", err.Error())
		return
	}
	fmt.Printf("Process '%s' added successfully.\n", name)
}

func (cli *CLI) startProcess(params []string) {
	if len(params) < 1 {
		fmt.Println("Usage: start <process_name> [args...]")
		return
	}
	name := params[0]
	proc, err := cli.manager.GetProcessByName(name)
	if err != nil {
		fmt.Println("Error:", err.Error())
		return
	}

	if err := proc.Start(params[1:]...); err != nil {
		fmt.Println("Error starting process:", err.Error())
		return
	}
	fmt.Printf("Process '%s' started with PID %d.\n", proc.Name, proc.Pid)
}

func (cli *CLI) stopProcess(params []string) {
	if len(params) < 1 {
		fmt.Println("Usage: stop <process_name>")
		return
	}
	name := params[0]
	proc, err := cli.manager.GetProcessByName(name)
	if err != nil {
		fmt.Println("Error:", err.Error())
		return
	}

	if err := proc.Stop(); err != nil {
		fmt.Println("Error stopping process:", err.Error())
		return
	}
	fmt.Printf("Process '%s' stopped.\n", proc.Name)
}

func (cli *CLI) listProcesses() {
	processes := cli.manager.Processes
	if len(processes) == 0 {
		fmt.Println("No processes are being managed.")
		return
	}
	fmt.Println("--- Managed Processes ---")
	for _, p := range processes {
		fmt.Printf("Name: %-15s | PID: %-7d | Status: %-10s | Schedule: %d\n", p.Name, p.Pid, p.GetStatus(), p.Schedul)
	}
}

func (cli *CLI) showStatus(params []string) {
	if len(params) < 1 {
		fmt.Println("Usage: status <process_name>")
		return
	}
	name := params[0]
	proc, err := cli.manager.GetProcessByName(name)
	if err != nil {
		fmt.Println("Error:", err.Error())
		return
	}
	fmt.Printf("--- Status for '%s' ---\n", name)
	fmt.Printf("  PID: %d\n", proc.Pid)
	fmt.Printf("  Path: %s\n", proc.Path)
	fmt.Printf("  Status: %s\n", proc.GetStatus())
	fmt.Printf("  Scheduling: %d\n", proc.Schedul)
	if proc.Timing != nil {
		fmt.Printf("  Scheduled Time: %s\n", proc.Timing.ScheduleTime.Format(time.RFC1123))
	}
}

func (cli *CLI) removeProcess(params []string) {
	if len(params) < 1 {
		fmt.Println("Usage: remove <process_name>")
		return
	}
	name := params[0]
	if err := cli.manager.RemoveProcess(name); err != nil {
		fmt.Println("Error removing process:", err.Error())
		return
	}
	fmt.Printf("Process '%s' has been removed.\n", name)
}

func showHelp() {
	fmt.Println("--- ExeProcessManager Help ---")
	fmt.Println("  help                            - Show this help message")
	fmt.Println("  list                            - List all managed processes")
	fmt.Println("  add <name> <path> <schedul>     - Add a new process (0=manual, 1=auto)")
	fmt.Println("  start <name> [args...]          - Start a manual process by name")
	fmt.Println("  stop <name>                     - Stop a running process by name")
	fmt.Println("  status <name>                   - Show detailed status of a process")
	fmt.Println("  remove <name>                   - Stop and remove a process from management")
	fmt.Println("  exit                            - (Deprecated) Use Ctrl+C to shut down gracefully")
	fmt.Println("--- Scheduling ---")
	fmt.Println("  createrule <rule_name> <time>   - Create a timing rule (time is Unix timestamp or RFC1123)")
	fmt.Println("  setjob <proc_name> <rule_name>  - Assign a timing rule to a process")
	fmt.Println("  startjob <proc_name>            - Start a scheduled process (will wait if needed)")
}