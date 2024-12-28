package api

import (
	"ExeProcessManager/process" // Adjust the import path to your actual package location
	"encoding/json"
	"fmt"
	"net/http"
)

// ProcessManager API
type ProcessAPI struct {
	Manager *process.ProcessManager
}

// NewProcessAPI creates a new ProcessAPI instance
func NewProcessAPI(pm *process.ProcessManager) *ProcessAPI {
	return &ProcessAPI{
		Manager: pm,
	}
}

// ListProcesses handles GET requests to list all processes
func (api *ProcessAPI) ListProcesses(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var response []map[string]interface{}
	for _, p := range api.Manager.Processes {
		response = append(response, map[string]interface{}{
			"Name":   p.Name,
			"PID":    p.Pid,
			"Status": p.GetStatus(),
			"Path":   p.Path,
		})
	}
	json.NewEncoder(w).Encode(response)
}

// StartProcess handles POST requests to start a process
func (api *ProcessAPI) StartProcess(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name string   `json:"name"`
		Args []string `json:"args"`
	}
	// Decode JSON body to get the process name and args
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Search for the process by name
	process, err := api.Manager.GetProcess(0, request.Name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Process %s not found", request.Name), http.StatusNotFound)
		return
	}

	// Start the process
	err = process.StartProcess(request.Args...)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to start process: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Process %s started successfully", request.Name)
}

// StopProcess handles POST requests to stop a process
func (api *ProcessAPI) StopProcess(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name string `json:"name"`
	}
	// Decode JSON body to get the process name
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Search for the process by name
	process, err := api.Manager.GetProcess(0, request.Name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Process %s not found", request.Name), http.StatusNotFound)
		return
	}

	// Stop the process
	err = process.Stop()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to stop process: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Process %s stopped successfully", request.Name)
}

// RestartProcess handles POST requests to restart a process
func (api *ProcessAPI) RestartProcess(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name string `json:"name"`
	}
	// Decode JSON body to get the process name
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Search for the process by name
	process, err := api.Manager.GetProcess(0, request.Name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Process %s not found", request.Name), http.StatusNotFound)
		return
	}

	// Restart the process
	err = process.Restart()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to restart process: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Process %s restarted successfully", request.Name)
}

// KillProcess handles POST requests to kill a process by its PID
func (api *ProcessAPI) KillProcess(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PID int `json:"pid"`
	}
	// Decode JSON body to get the PID
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Kill the process by PID
	err := api.Manager.KillProcess(request.PID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to kill process: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Process with PID %d killed successfully", request.PID)
}

// AddProcess handles POST requests to add a new process to the manager
func (api *ProcessAPI) AddProcess(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name    string `json:"name"`
		Path    string `json:"path"`
		Schedul int    `json:"schedul"`
	}
	// Decode JSON body to get the process information
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Add the process to the manager
	err := api.Manager.AddProcess(request.Name, request.Path, request.Schedul)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to add process: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Process %s added successfully", request.Name)
}

func (api *ProcessAPI) CreateTimingRule(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Time  string `json:"Time"`
		Tname string `json:"Name"` // Fixed the typo here
	}
	// Decode JSON body to get the process information
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	// Ensure the method exists in the process package
	err := process.CreateTimingRule(request.Tname, request.Time)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create timing rule: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// DeleteJob handles POST requests to delete a scheduled job for a process
func (api *ProcessAPI) DeleteJob(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name string `json:"name"`
	}
	// Decode JSON body to get the process name
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Search for the process by name
	_, err := api.Manager.GetProcess(0, request.Name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Process %s not found", request.Name), http.StatusNotFound)
		return
	}

	// Delete the job for the process
	err = api.Manager.DeleteJob(0, request.Name) // Assuming DeleteJob is a method of the Process struct
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete job for process: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Job for process %s deleted successfully", request.Name)
}

func (api *ProcessAPI) SetJob(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Time string `json:"Time"`
		Name string `json:"Name"`
	}
	// Decode JSON body to get the process information
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Search for the process by name
	process, err := api.Manager.GetProcess(0, request.Name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Process %s not found", request.Name), http.StatusNotFound)
		return
	}

	// Set the job for the process
	err = process.SetJob(request.Time)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to set job for process: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Job for process %s set successfully at time %s", request.Name, request.Time)
}

// StartJob handles POST requests to start a scheduled job for a process
func (api *ProcessAPI) StartJob(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name string `json:"name"`
	}
	// Decode JSON body to get the process name
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Search for the process by name
	process, err := api.Manager.GetProcess(0, request.Name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Process %s not found", request.Name), http.StatusNotFound)
		return
	}

	// Start the scheduled job
	err = process.StartJob()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to start job for process: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Job for process %s started successfully", request.Name)
}
