package api

import (
	"ExeProcessManager/config"
	"ExeProcessManager/process"
	"encoding/json"
	"log/slog"
	"net/http"
)

// ProcessAPI holds dependencies for the API handlers.
type ProcessAPI struct {
	Manager *process.ProcessManager
	Logger  *slog.Logger
	Config  *config.Config
}

// NewProcessAPI creates a new API handler instance.
func NewProcessAPI(pm *process.ProcessManager, logger *slog.Logger, cfg *config.Config) *ProcessAPI {
	return &ProcessAPI{
		Manager: pm,
		Logger:  logger,
		Config:  cfg,
	}
}

// Routes sets up all the API routes and returns an http.Handler.
// It now chains the authentication middleware with the logger middleware.
func (api *ProcessAPI) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /processes", api.listProcesses)
	mux.HandleFunc("POST /processes/add", api.addProcess)
	mux.HandleFunc("POST /processes/start", api.startProcess)
	mux.HandleFunc("POST /processes/stop", api.stopProcess)

	// Chain the middlewares: the request first hits the logger, then authentication.
	// You can reverse the order if you prefer.
	var handler http.Handler = mux
	handler = api.authMiddleware(handler)
	handler = api.logRequests(handler)

	return handler
}

// --- Handlers (No changes below this line) ---

func (api *ProcessAPI) listProcesses(w http.ResponseWriter, r *http.Request) {
	procs := api.Manager.Processes
	response := make([]map[string]interface{}, len(procs))
	for i, p := range procs {
		response[i] = map[string]interface{}{
			"name":   p.Name,
			"pid":    p.Pid,
			"status": p.GetStatus(),
			"path":   p.Path,
		}
	}
	respondWithJSON(w, http.StatusOK, response)
}

func (api *ProcessAPI) addProcess(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		Path    string `json:"path"`
		Schedul int    `json:"schedul"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	proc, err := api.Manager.AddProcess(req.Name, req.Path, req.Schedul)
	if err != nil {
		respondWithError(w, http.StatusConflict, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, proc)
}

func (api *ProcessAPI) startProcess(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string   `json:"name"`
		Args []string `json:"args"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	proc, err := api.Manager.GetProcessByName(req.Name)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	if err := proc.Start(req.Args...); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "process started"})
}

func (api *ProcessAPI) stopProcess(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	proc, err := api.Manager.GetProcessByName(req.Name)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	if err := proc.Stop(); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "process stopped"})
}

// --- Helper Functions (No changes here) ---

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal JSON response", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal Server Error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}