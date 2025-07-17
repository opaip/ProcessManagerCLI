#!/bin/bash

# run_tests.sh - A script to run all tests, check coverage, and display the report.

# --- Configuration ---
# Exit immediately if a command exits with a non-zero status.
set -e

# --- Main Logic ---
echo "Running all tests..."
# The -v flag provides verbose output, showing the status of each test.
go test -v ./...

echo ""
echo "Calculating test coverage..."
# The -coverprofile flag generates a file with coverage statistics.
go test -coverprofile=coverage.out ./...

echo ""
echo "Analyzing test coverage..."
# This command provides a summary of coverage by function.
go tool cover -func=coverage.out

echo ""
echo "To view the detailed HTML report, a web server will be started."
echo "Opening coverage report in your browser..."
echo "Press Ctrl+C in this terminal window to stop the server when you are done."

# This command starts a local web server to display the interactive coverage report.
go tool cover -html=coverage.out
