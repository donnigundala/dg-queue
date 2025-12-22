package dgqueue

import "fmt"

// logInfo logs an informational message.
func (m *Manager) logInfo(msg string, args ...interface{}) {
	if m.config.Logger != nil {
		// Prepend component name
		fullArgs := append([]interface{}{"component", "queue"}, args...)
		m.config.Logger.Info(msg, fullArgs...)
	} else {
		// Fallback to fmt.Printf for backward compatibility/development
		// But only if it's a significant event like start/stop
		// We don't want to spam stdout for every job
		if msg == "Queue manager starting" || msg == "Queue manager started" ||
			msg == "Queue manager stopping" || msg == "Queue manager stopped" {
			fmt.Printf("[Queue] %s", msg)
			for i := 0; i < len(args); i += 2 {
				if i+1 < len(args) {
					fmt.Printf(" %v=%v", args[i], args[i+1])
				}
			}
			fmt.Println()
		}
	}
}

// logError logs an error message.
func (m *Manager) logError(msg string, err error, args ...interface{}) {
	if m.config.Logger != nil {
		// Prepend component name and error
		newArgs := append([]interface{}{"component", "queue", "error", err}, args...)
		m.config.Logger.Error(msg, newArgs...)
	} else {
		// Fallback to fmt.Printf
		fmt.Printf("[Queue] ERROR: %s: %v", msg, err)
		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				fmt.Printf(" %v=%v", args[i], args[i+1])
			}
		}
		fmt.Println()
	}
}
