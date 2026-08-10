package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

func ParseLogLevel(val string) (LogLevel, error) {
	val = strings.ToUpper(val)
	logLevel := LogLevel(val)

	switch logLevel {
	case LevelDebug, LevelInfo, LevelWarn, LevelError:
		return logLevel, nil
	default:
		return "", fmt.Errorf("unknown log level: %q", val)
	}
}

type LogFormat string

const (
	FormatText LogFormat = "text"
	FormatJSON LogFormat = "json"
)

func ParseLogFormat(val string) (LogFormat, error) {
	val = strings.ToLower(val)
	logFormat := LogFormat(val)

	switch logFormat {
	case FormatText, FormatJSON:
		return logFormat, nil
	default:
		return "", fmt.Errorf("unknown log format: %q", logFormat)
	}
}

type Config struct {
	LoggerConfig   LoggerConfig
	HTTPConfig     HTTPConfig
	DatabaseConfig DatabaseConfig
}

type HTTPConfig struct {
	Port              int
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
}

type LoggerConfig struct {
	Level  LogLevel
	Format LogFormat
}

type DatabaseConfig struct {
	User     string
	Password string
	Host     string
	Port     int
	Name     string
	SSLMode  SSLMode
}

type SSLMode string

const (
	SSLRequire    SSLMode = "require"
	SSLDisable    SSLMode = "disable"
	SSLAllow      SSLMode = "allow"
	SSLPrefer     SSLMode = "prefer"
	SSLVerifyCA   SSLMode = "verify-ca"
	SSLVerifyFull SSLMode = "verify-full"
)

func ParseSSLMode(val string) (SSLMode, error) {
	val = strings.ToLower(val)
	sslMode := SSLMode(val)

	switch sslMode {
	case SSLRequire, SSLDisable, SSLAllow, SSLPrefer, SSLVerifyCA, SSLVerifyFull:
		return sslMode, nil
	default:
		return "", fmt.Errorf("unknown db ssl mode format: %q", sslMode)
	}
}

func (dc *DatabaseConfig) DSN() string {
	dbHostAndPort := fmt.Sprintf("%s:%d", dc.Host, dc.Port)

	dsnQuery := make(url.Values)
	dsnQuery.Set("sslmode", string(dc.SSLMode))

	dsnUrl := &url.URL{
		Scheme:   "postgresql",
		User:     url.UserPassword(dc.User, dc.Password),
		Host:     dbHostAndPort,
		Path:     dc.Name,
		RawQuery: dsnQuery.Encode(),
	}

	return dsnUrl.String()
}

func envStr(key, def string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return def
	}

	return val
}

func envInt(key string, def int) (int, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return def, nil
	}

	valInt, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("parse %s: invalid int %q: %w", key, val, err)
	}

	return valInt, nil
}

func envDuration(key string, def time.Duration) (time.Duration, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return def, nil
	}

	valDuration, err := time.ParseDuration(val)
	if err != nil {
		return 0, fmt.Errorf("parse %s: invalid duration %q: %w", key, val, err)
	}

	return valDuration, nil
}

func (c *Config) validate() []error {
	var validationErrors []error

	if c.HTTPConfig.Port > 65535 || c.HTTPConfig.Port < 1 {
		validationErrors = append(validationErrors, fmt.Errorf("port must be in 1...65535"))
	}

	if c.HTTPConfig.ReadHeaderTimeout < time.Second {
		validationErrors = append(validationErrors, fmt.Errorf("http.readHeaderTimeout less than a second"))
	}

	if c.HTTPConfig.ReadTimeout < time.Second {
		validationErrors = append(validationErrors, fmt.Errorf("http.readTimeout less than a second"))
	}

	if c.HTTPConfig.WriteTimeout < time.Second {
		validationErrors = append(validationErrors, fmt.Errorf("http.writeTimeout less than a second"))
	}

	if c.HTTPConfig.IdleTimeout < time.Second {
		validationErrors = append(validationErrors, fmt.Errorf("http.idleTimeout less than a second"))
	}

	if c.HTTPConfig.ShutdownTimeout < time.Second {
		validationErrors = append(validationErrors, fmt.Errorf("http.shutdownTimeout less than a second"))
	}

	if c.DatabaseConfig.User == "" {
		validationErrors = append(validationErrors, fmt.Errorf("db.user is unset"))
	}

	if c.DatabaseConfig.Password == "" {
		validationErrors = append(validationErrors, fmt.Errorf("db.password is unset"))
	}

	if c.DatabaseConfig.Host == "" {
		validationErrors = append(validationErrors, fmt.Errorf("db.host is unset"))
	}

	if c.DatabaseConfig.Port > 65535 || c.DatabaseConfig.Port < 1 {
		validationErrors = append(validationErrors, fmt.Errorf("db.port must be in 1...65535"))
	}

	if c.DatabaseConfig.Name == "" {
		validationErrors = append(validationErrors, fmt.Errorf("db.name is unset"))
	}

	_, err := ParseSSLMode(string(c.DatabaseConfig.SSLMode))
	if err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("unknown db.ssl format: %q", c.DatabaseConfig.SSLMode))
	}

	if c.DatabaseConfig.SSLMode == "" {
		validationErrors = append(validationErrors, fmt.Errorf("db.sslmode is unset"))
	}

	return validationErrors
}

func Load() (*Config, error) {
	var config Config
	var configErrors []error

	httpPort, err := envInt("HTTP_PORT", 8080)
	if err != nil {
		configErrors = append(configErrors, err)
	}

	logLevelRaw := envStr("LOG_LEVEL", "INFO")
	logLevel, err := ParseLogLevel(logLevelRaw)
	if err != nil {
		configErrors = append(configErrors, err)
	}

	logFormatRaw := envStr("LOG_FORMAT", "text")
	logFormat, err := ParseLogFormat(logFormatRaw)
	if err != nil {
		configErrors = append(configErrors, err)
	}

	loggerConfig := LoggerConfig{
		Level:  logLevel,
		Format: logFormat,
	}

	readHeaderTimeout, err := envDuration("HTTP_READ_HEADER_TIMEOUT", 10*time.Second)
	if err != nil {
		configErrors = append(configErrors, err)
	}

	readTimeout, err := envDuration("HTTP_READ_TIMEOUT", 10*time.Second)
	if err != nil {
		configErrors = append(configErrors, err)
	}

	writeTimeout, err := envDuration("HTTP_WRITE_TIMEOUT", 10*time.Second)
	if err != nil {
		configErrors = append(configErrors, err)
	}

	idleTimeout, err := envDuration("HTTP_IDLE_TIMEOUT", 10*time.Second)
	if err != nil {
		configErrors = append(configErrors, err)
	}

	shutdownTimeout, err := envDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		configErrors = append(configErrors, err)
	}

	httpConfig := HTTPConfig{
		Port:              httpPort,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ShutdownTimeout:   shutdownTimeout,
	}

	dbUser := envStr("DB_USER", "")
	dbPass := envStr("DB_PASSWORD", "")
	dbHost := envStr("DB_HOST", "")
	dbName := envStr("DB_NAME", "")

	dbPort, err := envInt("DB_PORT", 5432)
	if err != nil {
		configErrors = append(configErrors, err)
	}

	dbSSLMode := envStr("DB_SSLMODE", "")

	dbConfig := DatabaseConfig{
		User:     dbUser,
		Password: dbPass,
		Host:     dbHost,
		Port:     dbPort,
		Name:     dbName,
		SSLMode:  SSLMode(dbSSLMode),
	}

	config = Config{
		LoggerConfig:   loggerConfig,
		HTTPConfig:     httpConfig,
		DatabaseConfig: dbConfig,
	}

	validationErr := config.validate()
	if len(validationErr) > 0 {
		configErrors = append(configErrors, errors.Join(validationErr...))
	}

	if len(configErrors) > 0 {
		return nil, errors.Join(configErrors...)
	}

	return &config, nil
}
