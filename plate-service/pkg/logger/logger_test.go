package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected slog.Level
	}{
		{"debug lowercase", "debug", slog.LevelDebug},
		{"debug uppercase", "DEBUG", slog.LevelDebug},
		{"info lowercase", "info", slog.LevelInfo},
		{"info uppercase", "INFO", slog.LevelInfo},
		{"warn lowercase", "warn", slog.LevelWarn},
		{"warn uppercase", "WARN", slog.LevelWarn},
		{"warning lowercase", "warning", slog.LevelWarn},
		{"warning uppercase", "WARNING", slog.LevelWarn},
		{"error lowercase", "error", slog.LevelError},
		{"error uppercase", "ERROR", slog.LevelError},
		{"invalid level", "invalid", slog.LevelInfo},
		{"empty string", "", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNew_JSONFormat(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg := Config{
		Level:  "info",
		Format: "json",
	}

	logger := New(cfg)
	logger.Info("test message", "key", "value")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Output is not valid JSON: %v\nOutput: %s", err, output)
	}

	if logEntry["msg"] != "test message" {
		t.Errorf("Expected msg='test message', got %v", logEntry["msg"])
	}
	if logEntry["key"] != "value" {
		t.Errorf("Expected key='value', got %v", logEntry["key"])
	}
	if logEntry["level"] != "INFO" {
		t.Errorf("Expected level='INFO', got %v", logEntry["level"])
	}
}

func TestNew_TextFormat(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg := Config{
		Level:  "info",
		Format: "text",
	}

	logger := New(cfg)
	logger.Info("test message", "key", "value")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err == nil {
		t.Fatal("Output should not be valid JSON for text format")
	}

	if !strings.Contains(output, "test message") {
		t.Errorf("Output should contain 'test message', got: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("Output should contain 'key=value', got: %s", output)
	}
	if !strings.Contains(output, "INFO") {
		t.Errorf("Output should contain 'INFO', got: %s", output)
	}
}

func TestNew_LogLevels(t *testing.T) {
	tests := []struct {
		name        string
		configLevel string
		logLevel    string
		shouldLog   bool
	}{
		{"debug config logs debug", "debug", "debug", true},
		{"debug config logs info", "debug", "info", true},
		{"debug config logs warn", "debug", "warn", true},
		{"debug config logs error", "debug", "error", true},
		{"info config skips debug", "info", "debug", false},
		{"info config logs info", "info", "info", true},
		{"info config logs warn", "info", "warn", true},
		{"info config logs error", "info", "error", true},
		{"warn config skips debug", "warn", "debug", false},
		{"warn config skips info", "warn", "info", false},
		{"warn config logs warn", "warn", "warn", true},
		{"warn config logs error", "warn", "error", true},
		{"error config skips debug", "error", "debug", false},
		{"error config skips info", "error", "info", false},
		{"error config skips warn", "error", "warn", false},
		{"error config logs error", "error", "error", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			cfg := Config{
				Level:  tt.configLevel,
				Format: "text",
			}

			logger := New(cfg)

			// Log at the specified level
			switch tt.logLevel {
			case "debug":
				logger.Debug("test message")
			case "info":
				logger.Info("test message")
			case "warn":
				logger.Warn("test message")
			case "error":
				logger.Error("test message")
			}

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			hasOutput := len(output) > 0 && strings.Contains(output, "test message")

			if tt.shouldLog && !hasOutput {
				t.Errorf("Expected log output but got none. Config level: %s, Log level: %s", tt.configLevel, tt.logLevel)
			}
			if !tt.shouldLog && hasOutput {
				t.Errorf("Expected no log output but got: %s. Config level: %s, Log level: %s", output, tt.configLevel, tt.logLevel)
			}
		})
	}
}

func TestNew_DefaultFormat(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg := Config{
		Level:  "info",
		Format: "",
	}

	logger := New(cfg)
	logger.Info("test message")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Ожидает text
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err == nil {
		t.Fatal("Output should not be valid JSON for default (text) format")
	}

	if !strings.Contains(output, "test message") {
		t.Errorf("Output should contain 'test message', got: %s", output)
	}
}

func TestNew_DefaultLevel(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg := Config{
		Level:  "",
		Format: "text",
	}

	logger := New(cfg)

	logger.Debug("debug message")

	logger.Info("info message")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if strings.Contains(output, "debug message") {
		t.Error("Debug message should not be logged with default (info) level")
	}
	if !strings.Contains(output, "info message") {
		t.Error("Info message should be logged with default (info) level")
	}
}

func TestContextAttrs(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(contextHandler{slog.NewJSONHandler(&buf, nil)}).With("provider", "autonomera")

	ctx := WithAttrs(context.Background(), slog.String("job_id", "42"))
	nested := WithAttrs(ctx, slog.String("queue", "autonomera"))

	log.InfoContext(nested, "nested")
	log.InfoContext(ctx, "parent")
	log.Info("plain")

	var lines []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("unmarshal %q: %v", line, err)
		}
		lines = append(lines, entry)
	}

	if lines[0]["job_id"] != "42" || lines[0]["queue"] != "autonomera" || lines[0]["provider"] != "autonomera" {
		t.Errorf("nested context attrs missing: %v", lines[0])
	}
	if _, ok := lines[1]["queue"]; ok || lines[1]["job_id"] != "42" {
		t.Errorf("parent context must keep only its own attrs: %v", lines[1])
	}
	if _, ok := lines[2]["job_id"]; ok {
		t.Errorf("log without context must not have job attrs: %v", lines[2])
	}
}
