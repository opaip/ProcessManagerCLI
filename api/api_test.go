package api

import (
	"ExeProcessManager/config"
	"ExeProcessManager/process"
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// setupAPITest is a helper function to create all necessary components for an API test.
func setupAPITest(t *testing.T) (*ProcessAPI, *process.ProcessManager) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil)) // Discard logs during tests
	cfg := &config.Config{
		DataDir:     t.TempDir(),
		ScheduleDir: t.TempDir(),
	}
	pm := process.NewProcessManager(logger, cfg)
	api := NewProcessAPI(pm, logger, cfg)
	return api, pm
}

// TestListProcessesHandler tests the GET /processes endpoint.
func TestListProcessesHandler(t *testing.T) {
	api, pm := setupAPITest(t)

	// Add a process to the manager so the list is not empty
	_, _ = pm.AddProcess("test-proc", "/bin/sleep", 0)

	// Create a new HTTP request and a recorder to capture the response
	req := httptest.NewRequest(http.MethodGet, "/processes", nil)
	rr := httptest.NewRecorder()

	// Serve the request using the main router to include middleware
	api.Routes().ServeHTTP(rr, req)

	// Assert the status code is OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Assert the response body contains the process we added
	responseBody := rr.Body.String()
	if !strings.Contains(responseBody, "test-proc") {
		t.Errorf("handler response body does not contain the process name: got %s", responseBody)
	}
}

// TestAddProcessHandler tests the POST /processes/add endpoint.
func TestAddProcessHandler(t *testing.T) {
	api, manager := setupAPITest(t)

	payload := `{"name":"api-proc","path":"/bin/echo","schedul":0}`
	req := httptest.NewRequest(http.MethodPost, "/processes/add", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	api.Routes().ServeHTTP(rr, req)

	// Assert the status code is 201 Created
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	// Assert the process was actually added to the manager
	_, err := manager.GetProcessByName("api-proc")
	if err != nil {
		t.Errorf("process was not added to the manager after API call: %v", err)
	}

	// Assert the response body contains the new process name
	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("could not decode response body: %v", err)
	}
	if name, ok := response["name"]; !ok || name != "api-proc" {
		t.Errorf("response body does not have the correct name: got %v", response)
	}
}