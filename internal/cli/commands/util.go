package commands

import (
	"fmt"
	"time"
)

func defaultRunID() string {
	return fmt.Sprintf("%s", time.Now().UTC().Format("20060102-150405"))
}
