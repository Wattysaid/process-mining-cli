package logging

import (
	"log"
)

// Init configures logging output. Placeholder for structured logger.
func Init(level string, jsonOutput bool) error {
	_ = level
	_ = jsonOutput
	log.SetFlags(log.LstdFlags | log.LUTC)
	return nil
}
