ExeProcessManager - Advanced Process Management in Go
<p align="center">
<img src="https://placehold.co/600x300/1e293b/ffffff?text=ExeProcessManager&font=raleway" alt="Project Banner">
</p>

<p align="center">
<a href="#"><img src="https://img.shields.io/badge/build-passing-brightgreen" alt="Build Status"></a>
<a href="#"><img src="https://img.shields.io/badge/coverage-85%25-blue" alt="Code Coverage"></a>
<a href="#"><img src="https://img.shields.io/badge/Go-1.21%2B-blue.svg" alt="Go Version"></a>
<a href="#"><img src="https://img.shields.io/badge/license-MIT-lightgrey.svg" alt="License"></a>
</p>

ExeProcessManager is a powerful and lightweight tool for managing and monitoring system processes, written in Go. This project allows you to run your processes manually or on a schedule, view their status, and interact with them through a Command-Line Interface (CLI) and a secure REST API.

✨ Features
Full Process Lifecycle Management: Add, start, stop, remove, and view the status of processes.

State Persistence: The state of all processes is saved to disk, ensuring no data is lost after an application restart.

Scheduling: Define timing rules to automatically execute processes at a future time.

Dual Interface:

Command-Line Interface (CLI): For direct and fast management from the terminal.

REST API: For integration with other services and remote management.

API Security: All API routes are protected using secret API Keys.

Easy Configuration: All application settings are managed through a single config.json file.

Structured Logging: All events are logged in a standard format for easy debugging and monitoring.

Graceful Shutdown: The application listens for system signals (like Ctrl+C) and shuts down gracefully to ensure data integrity.

Comprehensive Tests: High test coverage to ensure the stability and correctness of the core application.

🚀 Getting Started
Follow these steps to set up and run the project.

Prerequisites
Go version 1.21 or higher must be installed.

Installation & Setup
Clone the repository:

git clone <your-repository-url>
cd ExeProcessManager

Configure the application:
Create a copy of the config.json file and modify its values according to your needs.

{
  "data_directory": "./data",
  "schedule_directory": "./schu",
  "log_level": "info",
  "api_listen_address": ":8080",
  "api_keys": [
    "your-secret-api-key-1",
    "another-secure-key-for-admin"
  ]
}

Important: Replace the api_keys with your own secure, randomly generated keys.

Build the project:

go build -o exepm .

This command creates an executable file named exepm (exepm.exe on Windows).

Run the application:

./exepm

Running this command will activate both the CLI and the API server.

🛠️ Usage
Command-Line Interface (CLI)
After running the application, you can enter the following commands in your terminal:

Command

Description

help

Show the list of all available commands.

list

List all managed processes.

add <name> <path> <sch>

Add a new process (sch: 0=manual, 1=auto).

start <name> [args...]

Start a manual process by its name.

stop <name>

Stop a running process.

status <name>

Show the detailed status of a process.

remove <name>

Completely remove a process from the manager.

REST API
All requests to the API must include the X-API-KEY header with a valid key.

Example: Get the list of processes with curl

# Set your API key in this variable
API_KEY="your-secret-api-key-1"

curl -H "X-API-KEY: $API_KEY" http://localhost:8080/processes

Main API Endpoints:

Method

Path

Request Body (JSON)

Description

GET

/processes

-

Get the list of all processes.

POST

/processes/add

{"name": "...", "path": "...", "schedul": 0}

Add a new process.

POST

/processes/start

{"name": "...", "args": ["..."]}

Start a process.

POST

/processes/stop

{"name": "..."}

Stop a process.

✅ Running Tests
To ensure all parts of the project are working correctly, you can run the unit tests:

go test ./...

🔮 Future Work
Resource Monitoring: Add the ability to monitor CPU and memory usage for each process.

Auto-Restart: Implement an auto-restart mechanism for processes that crash.

Advanced Scheduling: Support Cron-style scheduling rules.

Web UI: Build a web-based dashboard with React/Vue for graphical process management.

Log Management: Stream live logs from each process to the Web UI.

Notification System: Send alerts via Slack or Telegram on process failure.

🤝 Contributing
Contributions are welcome! Please feel free to submit a Pull Request or open an Issue on GitHub.

📜 License
This project is licensed under the MIT License. See the LICENSE file for more details.