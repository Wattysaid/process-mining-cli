package logging

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger struct {
	level       Level
	jsonOutput  bool
	writer      io.Writer
	baseWriter  io.Writer
	runLogFile  *os.File
	redactKeys  []string
	mu          sync.Mutex
	initialized bool
}

var global = &Logger{
	level:      LevelInfo,
	jsonOutput: false,
	writer:     os.Stdout,
	baseWriter: os.Stdout,
	redactKeys: []string{"password", "token", "secret", "api_key", "apikey", "credential", "key"},
}

// Init configures structured logging output.
func Init(level string, jsonOutput bool) error {
	parsed, err := parseLevel(level)
	if err != nil {
		return err
	}
	global.mu.Lock()
	defer global.mu.Unlock()
	global.level = parsed
	global.jsonOutput = jsonOutput
	global.baseWriter = os.Stdout
	global.writer = global.baseWriter
	global.initialized = true
	return nil
}

// SetRunLog routes log output to outputs/<run-id>/run.log in addition to stdout.
func SetRunLog(outputDir string) error {
	if outputDir == "" {
		return errors.New("output directory is required")
	}
	logPath := filepath.Join(outputDir, "run.log")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	global.mu.Lock()
	defer global.mu.Unlock()
	if global.runLogFile != nil {
		_ = global.runLogFile.Close()
	}
	global.runLogFile = file
	global.writer = io.MultiWriter(global.baseWriter, file)
	return nil
}

// CloseRunLog closes the run log file if open.
func CloseRunLog() {
	global.mu.Lock()
	defer global.mu.Unlock()
	if global.runLogFile != nil {
		_ = global.runLogFile.Close()
		global.runLogFile = nil
		global.writer = global.baseWriter
	}
}

func Debug(msg string, fields map[string]any) {
	global.log(LevelDebug, msg, fields)
}

func Info(msg string, fields map[string]any) {
	global.log(LevelInfo, msg, fields)
}

func Warn(msg string, fields map[string]any) {
	global.log(LevelWarn, msg, fields)
}

func Error(msg string, fields map[string]any) {
	global.log(LevelError, msg, fields)
}

func (l *Logger) log(level Level, msg string, fields map[string]any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.initialized {
		log.SetFlags(log.LstdFlags | log.LUTC)
	}
	if level < l.level {
		return
	}
	redacted := l.redactFields(fields)
	entry := map[string]any{
		"ts":    time.Now().UTC().Format(time.RFC3339),
		"level": levelName(level),
		"msg":   msg,
	}
	for key, value := range redacted {
		entry[key] = value
	}
	if l.jsonOutput {
		blob, err := json.Marshal(entry)
		if err != nil {
			fmt.Fprintln(l.writer, "{\"level\":\"error\",\"msg\":\"failed to encode log\"}")
			return
		}
		fmt.Fprintln(l.writer, string(blob))
		return
	}
	fmt.Fprintln(l.writer, formatLogfmt(entry))
}

func (l *Logger) redactFields(fields map[string]any) map[string]any {
	if len(fields) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(fields))
	for key, value := range fields {
		if l.shouldRedact(key) {
			out[key] = "[REDACTED]"
			continue
		}
		out[key] = value
	}
	return out
}

func (l *Logger) shouldRedact(key string) bool {
	normalized := strings.ToLower(key)
	for _, token := range l.redactKeys {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}

func parseLevel(level string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return LevelDebug, nil
	case "info", "":
		return LevelInfo, nil
	case "warn", "warning":
		return LevelWarn, nil
	case "error":
		return LevelError, nil
	default:
		return LevelInfo, fmt.Errorf("unknown log level: %s", level)
	}
}

func levelName(level Level) string {
	switch level {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	default:
		return "info"
	}
}

func formatLogfmt(entry map[string]any) string {
	keys := make([]string, 0, len(entry))
	for key := range entry {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(entry))
	for _, key := range keys {
		value := entry[key]
		parts = append(parts, fmt.Sprintf("%s=%q", key, fmt.Sprint(value)))
	}
	return strings.Join(parts, " ")
}
