package config

import (
	"os"
)

type Config struct {
	// App 設定
	AppName    string
	AppVersion string
	AppPort    string
	// OpenTelemetry 設定
	OtelHost              string
	OtelPort              string
	OtelServiceName       string
	OtelServiceVersion    string
	OtelScopeAPIName      string
	OtelScopeDBName       string
	OtelScopeCacheName    string
	OtelScopeRendererName string
}

func Load() *Config {
	return &Config{
		AppName:               getEnv("SERVICE_NAME", "myapp"),
		AppVersion:            getEnv("SERVICE_VERSION", "0.0.1"),
		AppPort:               getEnv("APP_PORT", "8080"),
		OtelHost:              getEnv("OTEL_HOST", "localhost"),
		OtelPort:              getEnv("OTEL_PORT", "4317"), // OTLP gRPC
		OtelServiceName:       getEnv("OTEL_SERVICE_NAME", "myapp-api"),
		OtelServiceVersion:    getEnv("OTEL_SERVICE_VERSION", "0.0.1"),
		OtelScopeAPIName:      getEnv("OTEL_SCOPE_API_NAME", "myapp/api"),
		OtelScopeDBName:       getEnv("OTEL_SCOPE_DB_NAME", "myapp/db"),
		OtelScopeCacheName:    getEnv("OTEL_SCOPE_CACHE_NAME", "myapp/cache"),
		OtelScopeRendererName: getEnv("OTEL_SCOPE_RENDERER_NAME", "myapp/renderer"),
	}
}

func getEnv(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}
