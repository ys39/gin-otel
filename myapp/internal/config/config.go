package config

import (
	"os"
)

type Config struct {
	// App 設定
	AppName    string
	AppVersion string
	AppPort    string
	// DB 設定
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	DBSystem     string
	DBMockSystem string
	// OpenTelemetry 設定
	OtelHost              string
	OtelPort              string
	OtelServiceName       string
	OtelServiceVersion    string
	OtelScopeAPIName      string
	OtelScopeDBName       string
	OtelScopeCacheName    string
	OtelScopeRendererName string
	OtelDeploymentEnv     string
	OtelSDKName           string
	OtelSDKLanguage       string
	OtelSDKVersion        string
}

func Load() *Config {
	return &Config{
		// App 設定
		AppName:    getEnv("SERVICE_NAME", "myapp"),
		AppVersion: getEnv("SERVICE_VERSION", "0.0.1"),
		AppPort:    getEnv("APP_PORT", "8080"),
		OtelHost:   getEnv("OTEL_HOST", "localhost"),
		// DB 設定
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "3306"),
		DBUser:       getEnv("DB_USER", "dbuser"),
		DBPassword:   getEnv("DB_PASSWORD", "**********"),
		DBName:       getEnv("DB_NAME", "myapp_db"),
		DBSystem:     getEnv("DB_SYSTEM", "mysql"),
		DBMockSystem: getEnv("DB_MOCK_SYSTEM", "mock"), // モックシステムの指定
		// OpenTelemetry 設定
		OtelPort:              getEnv("OTEL_PORT", "4317"), // OTLP gRPC
		OtelServiceName:       getEnv("OTEL_SERVICE_NAME", "myapp-api"),
		OtelServiceVersion:    getEnv("OTEL_SERVICE_VERSION", "0.0.1"),
		OtelScopeAPIName:      getEnv("OTEL_SCOPE_API_NAME", "myapp/api"),
		OtelScopeDBName:       getEnv("OTEL_SCOPE_DB_NAME", "myapp/db"),
		OtelScopeCacheName:    getEnv("OTEL_SCOPE_CACHE_NAME", "myapp/cache"),
		OtelScopeRendererName: getEnv("OTEL_SCOPE_RENDERER_NAME", "myapp/renderer"),
		OtelDeploymentEnv:     getEnv("OTEL_DEPLOYMENT_ENV", "development"),
		OtelSDKName:           getEnv("OTEL_SDK_NAME", "opentelemetry"),
		OtelSDKLanguage:       getEnv("OTEL_SDK_LANGUAGE", "go"),
		OtelSDKVersion:        getEnv("OTEL_SDK_VERSION", "1.24.0"),
	}
}

func getEnv(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}
