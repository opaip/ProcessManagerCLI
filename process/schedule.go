package process

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// TimingRule defines scheduling rules for a process.
type TimingRule struct {
	ScheduleTime time.Time `json:"schedule_time"`
}

// CreateTimingRule creates and saves a new timing rule.
func (pm *ProcessManager) CreateTimingRule(ruleName string, scheduleInput string) error {
	var scheduleTime time.Time
	var err error

	// Try parsing as Unix timestamp
	if timestamp, err := strconv.ParseInt(scheduleInput, 10, 64); err == nil {
		scheduleTime = time.Unix(timestamp, 0)
	} else {
		// Fallback to RFC1123 format
		scheduleTime, err = time.Parse(time.RFC1123, scheduleInput)
		if err != nil {
			return fmt.Errorf("invalid time format: must be Unix timestamp or RFC1123 string: %w", err)
		}
	}

	rule := TimingRule{
		ScheduleTime: scheduleTime,
	}

	filePath := filepath.Join(pm.config.ScheduleDir, "rules", ruleName+".json")
	if FileExists(filePath) {
		return fmt.Errorf("timing rule '%s' already exists", ruleName)
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for timing rules: %w", err)
	}

	if err := SaveToFile(filePath, rule); err != nil {
		return fmt.Errorf("failed to save timing rule file: %w", err)
	}

	pm.logger.Info("timing rule created successfully", "name", ruleName, "time", scheduleTime.Format(time.RFC1123))
	return nil
}

// SetJob assigns a timing rule to a process.
func (p *Process) SetJob(timingRuleName string) error {
	if p.Schedul != 1 {
		return fmt.Errorf("process '%s' is not configured for automatic scheduling", p.Name)
	}

	rule := TimingRule{}
	filePath := filepath.Join(p.manager.config.ScheduleDir, "rules", timingRuleName+".json")
	if err := LoadFromFile(filePath, &rule); err != nil {
		return fmt.Errorf("failed to load timing rule '%s': %w", timingRuleName, err)
	}

	p.Timing = &rule
	p.manager.logger.Info("job set for process", "name", p.Name, "time", p.Timing.ScheduleTime.Format(time.RFC1123))

	// Save the process state with the new timing information
	return p.SaveState()
}

// StartJob starts a process based on its schedule.
func (p *Process) StartJob() error {
	if p.Stat == 1 {
		p.manager.logger.Warn("cannot start job, process is already running", "name", p.Name)
		return nil // Not an error, just can't start again
	}

	if p.Timing == nil {
		return fmt.Errorf("process '%s' has no job timing configured", p.Name)
	}

	now := time.Now()
	scheduleTime := p.Timing.ScheduleTime

	if now.Before(scheduleTime) {
		waitDuration := time.Until(scheduleTime)
		p.manager.logger.Info("job scheduled for the future, waiting...", "name", p.Name, "duration", waitDuration.String())

		// Non-blocking wait
		time.AfterFunc(waitDuration, func() {
			if p.IsJobDeleted == 1 {
				p.manager.logger.Info("job was deleted before it could run", "name", p.Name)
				return
			}
			p.manager.logger.Info("scheduled time reached, starting process", "name", p.Name)
			if err := p.Start(); err != nil {
				p.manager.logger.Error("failed to auto-start scheduled process", "name", p.Name, "error", err)
			}
		})
		return nil
	}

	// If the time has already passed, start it immediately
	p.manager.logger.Info("job schedule is in the past, starting immediately", "name", p.Name)
	return p.Start()
}