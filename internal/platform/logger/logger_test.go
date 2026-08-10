package logger

import (
	"bytes"
	"encoding/json"
	"go-platform-template/internal/platform/config"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestTextLogger(t *testing.T) {
	var outputBuf bytes.Buffer

	// Text logger
	textConfig := config.LoggerConfig{
		Level:  config.LevelDebug,
		Format: config.FormatText,
	}

	textLogger := newWithWriter(&outputBuf, textConfig)

	// Write in buffer
	textLogger.Info("ABRACADABRA")

	outputString := outputBuf.String()

	if !strings.Contains(outputString, "ABRACADABRA") {
		t.Errorf("expected string in output not found. output: %q", outputBuf.String())
	}

	if !strings.Contains(outputString, "level=INFO") {
		t.Errorf("expected 'level=INFO' level in output: %q", outputBuf.String())
	}
}

func TestJSONLogger(t *testing.T) {
	var outputBuf bytes.Buffer

	// JSON logger
	jsonConfig := config.LoggerConfig{
		Level:  config.LevelInfo,
		Format: config.FormatJSON,
	}

	jsonLogger := newWithWriter(&outputBuf, jsonConfig)

	// Write in buffer
	jsonLogger.Error("Test error", "attribute", "TestAttr")

	var outputJson struct {
		Time      time.Time  `json:"time"`
		Level     slog.Level `json:"level"`
		Message   string     `json:"msg"`
		Attribute string     `json:"attribute"`
	}

	err := json.Unmarshal(outputBuf.Bytes(), &outputJson)
	if err != nil {
		t.Fatalf("failed to unmarshal buffer: %v", err)
	}

	if outputJson.Level != slog.LevelError {
		t.Errorf("expected level %q, got %q", slog.LevelError, outputJson.Level)
	}

	if outputJson.Message != "Test error" {
		t.Errorf("expected message 'Test error', got %q", outputJson.Message)
	}

	if outputJson.Attribute != "TestAttr" {
		t.Errorf("expected attribute value 'TestAttr', got %q", outputJson.Attribute)
	}
}

func TestZeroBufferOnOtherLevel(t *testing.T) {
	var outputBuf bytes.Buffer

	jsonConfig := config.LoggerConfig{
		Level:  config.LevelError,
		Format: config.FormatJSON,
	}

	jsonLogger := newWithWriter(&outputBuf, jsonConfig)

	// Write in buffer
	jsonLogger.Warn("Warning test")

	if outputBuf.String() != "" {
		t.Errorf("expected zero string, got: %s", outputBuf.String())
	}
}

func TestZeroConfig(t *testing.T) {
	var outputBuf bytes.Buffer

	// Assert that zero config got no panic
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("the test panicked with: %v", r)
		}
	}()

	zeroConfig := config.LoggerConfig{}

	logger := newWithWriter(&outputBuf, zeroConfig)

	logger.Warn("Something gonna happened")
}
