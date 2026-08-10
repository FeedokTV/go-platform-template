package config_test

import (
	"go-platform-template/internal/platform/config"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func clearConfigEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		"LOG_LEVEL", "LOG_FORMAT", "HTTP_PORT",
		"HTTP_READ_HEADER_TIMEOUT", "HTTP_READ_TIMEOUT",
		"HTTP_WRITE_TIMEOUT", "HTTP_IDLE_TIMEOUT",
		"DB_USER", "DB_PASSWORD", "DB_HOST", "DB_PORT", "DB_NAME", "DB_SSLMODE",
	}
	for _, k := range keys {
		t.Setenv(k, "")
		_ = os.Unsetenv(k)
	}
}

func TestLoadSuccessful(t *testing.T) {
	clearConfigEnv(t)
	fiveSecDuration := time.Duration(time.Second * 5)

	t.Setenv("LOG_LEVEL", "DEBUG")
	t.Setenv("LOG_FORMAT", "json")
	t.Setenv("HTTP_PORT", "8080")
	t.Setenv("HTTP_READ_TIMEOUT", "5s")
	t.Setenv("HTTP_READ_HEADER_TIMEOUT", "5s")
	t.Setenv("HTTP_WRITE_TIMEOUT", "5s")
	t.Setenv("HTTP_IDLE_TIMEOUT", "5s")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT", "5s")
	t.Setenv("DB_USER", "test")
	t.Setenv("DB_PASSWORD", "pass")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "6352")
	t.Setenv("DB_NAME", "testdb")
	t.Setenv("DB_SSLMODE", "require")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config load error: %v", err)
	}

	if cfg.LoggerConfig.Level != config.LevelDebug {
		t.Error("log level does not correspond")
	}

	if cfg.LoggerConfig.Format != config.FormatJSON {
		t.Error("log.format does not correspond")
	}

	if cfg.HTTPConfig.Port != 8080 {
		t.Error("http.port does not correspond")
	}

	if cfg.HTTPConfig.ReadHeaderTimeout != fiveSecDuration {
		t.Error("http.readHeaderTimeout does not correspond")
	}

	if cfg.HTTPConfig.ReadTimeout != fiveSecDuration {
		t.Error("http.readTimeout does not correspond")
	}

	if cfg.HTTPConfig.WriteTimeout != fiveSecDuration {
		t.Error("http.writetimeout does not correspond")
	}

	if cfg.HTTPConfig.IdleTimeout != fiveSecDuration {
		t.Error("http.idleTimeout does not correspond")
	}

	if cfg.HTTPConfig.ShutdownTimeout != fiveSecDuration {
		t.Error("http.shutdowntimeout does not correspond")
	}

	if cfg.DatabaseConfig.DSN() != "postgresql://test:pass@localhost:6352/testdb?sslmode=require" { // #nosec G101
		t.Error("database dsn does not correspond")
	}
}

func TestValidationsRules(t *testing.T) {
	clearConfigEnv(t)

	tests := []struct {
		Name string
		Want string
	}{
		{
			Name: "unknown log level",
			Want: "unknown log level: \"FOO\"",
		},
		{
			Name: "unknown log format",
			Want: "unknown log format: \"foo_foo\"",
		},
		{
			Name: "bad http.port",
			Want: "port must be in 1...65535",
		},
		{
			Name: "http.readtimeout duration less than a second",
			Want: "http.readTimeout less than a second",
		},
		{
			Name: "http.readheadertimeout duration less than a second",
			Want: "http.readHeaderTimeout less than a second",
		},
		{
			Name: "http.writetimeout duration less than a second",
			Want: "http.writeTimeout less than a second",
		},
		{
			Name: "http.idletimeout duration less than a second",
			Want: "http.idleTimeout less than a second",
		},
		{
			Name: "http.shutdowntimeout duration less than a second",
			Want: "http.shutdownTimeout less than a second",
		},
		{
			Name: "bad db.user",
			Want: "db.user is unset",
		},
		{
			Name: "bad db.password",
			Want: "db.password is unset",
		},
		{
			Name: "bad db.host",
			Want: "db.host is unset",
		},
		{
			Name: "bad db.port",
			Want: "db.port must be in 1...65535",
		},
		{
			Name: "bad db.name",
			Want: "db.name is unset",
		},
		{
			Name: "bad db.sslmode",
			Want: "unknown db.ssl format",
		},
	}

	t.Setenv("LOG_LEVEL", "FOO")
	t.Setenv("LOG_FORMAT", "FOO_FOO")
	t.Setenv("HTTP_PORT", "0")
	t.Setenv("HTTP_READ_TIMEOUT", "0s")
	t.Setenv("HTTP_READ_HEADER_TIMEOUT", "50ns")
	t.Setenv("HTTP_WRITE_TIMEOUT", "0s")
	t.Setenv("HTTP_IDLE_TIMEOUT", "0s")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT", "0s")
	t.Setenv("DB_PORT", "0")
	t.Setenv("DB_SSLMODE", "blabla")

	_, err := config.Load()
	if err == nil {
		t.Fatal("config loaded successfully with bad values")
	}

	validationErr := err.Error()
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			if !strings.Contains(validationErr, test.Want) {
				t.Errorf("expected error %v", test.Want)
			}
		})
	}

}

func TestConfigDefault(t *testing.T) {
	clearConfigEnv(t)

	tenSecDuration := time.Duration(10 * time.Second)

	t.Setenv("DB_USER", "test")
	t.Setenv("DB_PASSWORD", "pass")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "6352")
	t.Setenv("DB_NAME", "testdb")
	t.Setenv("DB_SSLMODE", "require")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config load error: %v", err)
	}

	if cfg.LoggerConfig.Level != config.LevelInfo {
		t.Errorf("expected %q, got %s", config.LevelInfo, cfg.LoggerConfig.Level)
	}

	if cfg.LoggerConfig.Format != config.FormatText {
		t.Errorf("expected %q, got %s", config.FormatText, cfg.LoggerConfig.Format)
	}

	if cfg.HTTPConfig.Port != 8080 {
		t.Errorf("expected %d, got %d", 8080, cfg.HTTPConfig.Port)
	}

	if cfg.HTTPConfig.ReadHeaderTimeout != tenSecDuration {
		t.Errorf("expected %q, got %s", tenSecDuration, cfg.HTTPConfig.ReadHeaderTimeout)
	}

	if cfg.HTTPConfig.ReadTimeout != tenSecDuration {
		t.Errorf("expected %q, got %s", tenSecDuration, cfg.HTTPConfig.ReadTimeout)
	}

	if cfg.HTTPConfig.WriteTimeout != tenSecDuration {
		t.Errorf("expected %q, got %s", tenSecDuration, cfg.HTTPConfig.WriteTimeout)
	}

	if cfg.HTTPConfig.IdleTimeout != tenSecDuration {
		t.Errorf("expected %q, got %s", tenSecDuration, cfg.HTTPConfig.IdleTimeout)
	}

	if cfg.HTTPConfig.ShutdownTimeout != tenSecDuration {
		t.Errorf("expected %q, got %s", tenSecDuration, cfg.HTTPConfig.ShutdownTimeout)
	}
}

func TestConfigDSNGeneration(t *testing.T) {
	clearConfigEnv(t)

	t.Setenv("DB_USER", "test")
	t.Setenv("DB_PASSWORD", "p@ss/w#rd")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "6352")
	t.Setenv("DB_NAME", "testdb")
	t.Setenv("DB_SSLMODE", "require")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config load error: %v", err)
	}

	parsedDsn, err := url.Parse(cfg.DatabaseConfig.DSN())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsedDsn.User.Username() != "test" {
		t.Error("database dsn.username does not correspond")
	}

	parsedPass, passSet := parsedDsn.User.Password()
	if !passSet {
		t.Error("database dsn.password is not set")

	}
	if parsedPass != "p@ss/w#rd" {
		t.Error("database dsn.password does not correspond")
	}

	if parsedDsn.Host != "localhost:6352" {
		t.Error("database dsn.host does not correspond")
	}

	if parsedDsn.Path != "/testdb" {
		t.Error("database dsn.name does not correspond")
	}
	if parsedDsn.Query().Encode() != "sslmode=require" {
		t.Error("database dsn.sslmode does not correspond")
	}
}
