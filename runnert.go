package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// main is the entry point for the test runner application.
func main() {
	fmt.Println("🚀 Starting comprehensive test runner...")

	// Define the filenames for coverage artifacts.
	coverageProfile := "coverage.out"
	coverageHTML := "coverage.html"

	// Clean up previous artifacts before starting.
	cleanup(coverageProfile, coverageHTML)
	// Ensure cleanup is also run on exit.
	defer cleanup(coverageProfile, coverageHTML)

	// Step 1: Run all tests with verbose output.
	if err := runTests(); err != nil {
		log.Fatalf("❌ Tests failed: %v", err)
	}
	fmt.Println("\n✅ All tests passed successfully.")

	// Step 2: Generate the coverage profile.
	if err := generateCoverageProfile(coverageProfile); err != nil {
		log.Fatalf("❌ Failed to generate coverage profile: %v", err)
	}
	fmt.Printf("\n✅ Coverage profile generated at '%s'.\n", coverageProfile)

	// Step 3: Display function-level coverage summary in the console.
	if err := showCoverageSummary(coverageProfile); err != nil {
		log.Fatalf("❌ Failed to analyze coverage: %v", err)
	}

	// Step 4: Generate a static HTML report from the coverage profile.
	if err := generateHTMLReport(coverageProfile, coverageHTML); err != nil {
		log.Fatalf("❌ Failed to generate HTML report: %v", err)
	}
	fmt.Printf("\n✅ HTML coverage report generated at '%s'.\n", coverageHTML)

	// Step 5: Open the HTML report in the default browser.
	if err := openInBrowser(coverageHTML); err != nil {
		log.Printf("⚠️ Could not open report in browser: %v", err)
		fmt.Println("Please open the 'coverage.html' file manually.")
	} else {
		fmt.Println("\n✅ Opening coverage report in your browser...")
	}

	fmt.Println("\n🎉 Test run complete.")
}

// runTests executes `go test -v ./...`.
func runTests() error {
	fmt.Println("\n--- Running All Tests ---")
	return executeCommand("go", "test", "-v", "./...")
}

// generateCoverageProfile executes `go test -coverprofile=... ./...`.
func generateCoverageProfile(outFile string) error {
	fmt.Println("\n--- Generating Coverage Profile ---")
	return executeCommand("go", "test", fmt.Sprintf("-coverprofile=%s", outFile), "./...")
}

// showCoverageSummary executes `go tool cover -func=...`.
func showCoverageSummary(inFile string) error {
	fmt.Println("\n--- Coverage Summary by Function ---")
	return executeCommand("go", "tool", "cover", fmt.Sprintf("-func=%s", inFile))
}

// generateHTMLReport executes `go tool cover -html=... -o ...`.
func generateHTMLReport(profileFile, htmlFile string) error {
	return executeCommand("go", "tool", "cover", fmt.Sprintf("-html=%s", profileFile), "-o", htmlFile)
}

// openInBrowser opens the specified file path in the user's default web browser.
func openInBrowser(path string) error {
	var cmd *exec.Cmd
	fullPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", fullPath)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", fullPath)
	case "darwin":
		cmd = exec.Command("open", fullPath)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}

// executeCommand is a helper function to run an external command and stream its output.
func executeCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	// Connect the command's stdout and stderr to the main process's stdout and stderr.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// cleanup removes the generated artifact files.
func cleanup(files ...string) {
	fmt.Println("\n--- Cleaning up artifacts ---")
	for _, file := range files {
		err := os.Remove(file)
		if err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: could not remove file %s: %v", file, err)
		}
	}
}
