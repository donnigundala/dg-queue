package queue

import "fmt"

// logInfo logs an informational message.
func (m *Manager) logInfo(msg string, keysAndValues ...interface{}) {
	if m.config.Logger != nil {
		m.config.Logger.Info(msg, keysAndValues...)
	} else {
		// Fallback to fmt.Printf for backward compatibility/development
		// But only if it's a significant event like start/stop
		// We don't want to spam stdout for every job
		if msg == "Queue manager starting" || msg == "Queue manager started" ||
			msg == "Queue manager stopping" || msg == "Queue manager stopped" {
			fmt.Printf("[Queue] %s", msg)
			for i := 0; i < len(keysAndValues); i += 2 {
				if i+1 < len(keysAndValues) {
					fmt.Printf(" %v=%v", keysAndValues[i], keysAndValues[i+1])
				}
			}
			fmt.Println()
		}
	}
}

// logError logs an error message.
func (m *Manager) logError(msg string, err error, keysAndValues ...interface{}) {
	if m.config.Logger != nil {
		m.config.Logger.Error(msg, err, keysAndValues...)
	} else {
		// Fallback to fmt.Printf
		fmt.Printf("[Queue] ERROR: %s: %v", msg, err)
		for i := 0; i < len(keysAndValues); i += 2 {
			if i+1 < len(keysAndValues) {
				fmt.Printf(" %v=%v", keysAndValues[i], keysAndValues[i+1])
			}
		}
		fmt.Println()
	}
}
