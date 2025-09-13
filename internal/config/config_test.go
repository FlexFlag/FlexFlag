package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Defaults(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()
	
	// Clear environment variables that might override defaults in CI
	envVars := []string{
		"FLEXFLAG_SERVER_HOST",
		"FLEXFLAG_SERVER_PORT",
		"FLEXFLAG_DATABASE_HOST",
		"FLEXFLAG_DATABASE_PORT",
		"FLEXFLAG_DATABASE_USERNAME",
		"FLEXFLAG_DATABASE_PASSWORD", 
		"FLEXFLAG_DATABASE_DATABASE",
		"FLEXFLAG_REDIS_HOST",
		"FLEXFLAG_REDIS_PORT",
	}
	
	originalValues := make(map[string]string)
	for _, envVar := range envVars {
		originalValues[envVar] = os.Getenv(envVar)
		os.Unsetenv(envVar)
	}
	
	defer func() {
		// Restore original environment variables
		for envVar, originalValue := range originalValues {
			if originalValue != "" {
				os.Setenv(envVar, originalValue)
			}
		}
	}()
	
	config, err := Load()
	require.NoError(t, err)
	require.NotNil(t, config)

	// Test server defaults
	assert.Equal(t, "0.0.0.0", config.Server.Host)
	assert.Equal(t, 8080, config.Server.Port)
	assert.Equal(t, 30, config.Server.ReadTimeout)
	assert.Equal(t, 30, config.Server.WriteTimeout)
	assert.Equal(t, "development", config.Server.Environment)

	// Test database defaults
	assert.Equal(t, "localhost", config.Database.Host)
	assert.Equal(t, 5432, config.Database.Port)
	assert.Equal(t, "flexflag", config.Database.Username)
	assert.Equal(t, "flexflag", config.Database.Password)
	assert.Equal(t, "flexflag", config.Database.Database)
	assert.Equal(t, "disable", config.Database.SSLMode)
	assert.Equal(t, 10, config.Database.MaxConns)
	assert.Equal(t, 2, config.Database.MinConns)

	// Test redis defaults
	assert.Equal(t, "localhost", config.Redis.Host)
	assert.Equal(t, 6379, config.Redis.Port)
	assert.Equal(t, "", config.Redis.Password)
	assert.Equal(t, 0, config.Redis.Database)
	assert.Equal(t, 300, config.Redis.TTL)

	// Test auth defaults
	assert.Equal(t, "your-secret-key", config.Auth.JWTSecret)
	assert.Equal(t, 3600, config.Auth.TokenExpiry)
	assert.Equal(t, 86400, config.Auth.RefreshExpiry)

	// Test logging defaults
	assert.Equal(t, "info", config.Logging.Level)
	assert.Equal(t, "json", config.Logging.Format)
	assert.Equal(t, "stdout", config.Logging.Output)
}

func TestConfig_EnvironmentVariables(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()
	
	// Set environment variables
	defer func() {
		os.Unsetenv("FLEXFLAG_SERVER_HOST")
		os.Unsetenv("FLEXFLAG_SERVER_PORT")
		os.Unsetenv("FLEXFLAG_DATABASE_HOST")
		os.Unsetenv("FLEXFLAG_AUTH_JWT_SECRET")
	}()

	os.Setenv("FLEXFLAG_SERVER_HOST", "127.0.0.1")
	os.Setenv("FLEXFLAG_SERVER_PORT", "9090")
	os.Setenv("FLEXFLAG_DATABASE_HOST", "db.example.com")
	os.Setenv("FLEXFLAG_AUTH_JWT_SECRET", "test-secret")

	config, err := Load()
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify environment variables override defaults
	assert.Equal(t, "127.0.0.1", config.Server.Host)
	assert.Equal(t, 9090, config.Server.Port)
	assert.Equal(t, "db.example.com", config.Database.Host)
	assert.Equal(t, "test-secret", config.Auth.JWTSecret)
}

func TestConfig_SetDefaults(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()
	
	setDefaults()

	// Test that defaults are set correctly
	assert.Equal(t, "0.0.0.0", viper.GetString("server.host"))
	assert.Equal(t, 8080, viper.GetInt("server.port"))
	assert.Equal(t, "localhost", viper.GetString("database.host"))
	assert.Equal(t, 5432, viper.GetInt("database.port"))
	assert.Equal(t, "localhost", viper.GetString("redis.host"))
	assert.Equal(t, 6379, viper.GetInt("redis.port"))
	assert.Equal(t, "your-secret-key", viper.GetString("auth.jwt_secret"))
	assert.Equal(t, "info", viper.GetString("logging.level"))
}

func TestServerConfig_Struct(t *testing.T) {
	config := ServerConfig{
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  60,
		WriteTimeout: 60,
		Environment:  "production",
	}

	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 8080, config.Port)
	assert.Equal(t, 60, config.ReadTimeout)
	assert.Equal(t, 60, config.WriteTimeout)
	assert.Equal(t, "production", config.Environment)
}

func TestDatabaseConfig_Struct(t *testing.T) {
	config := DatabaseConfig{
		Host:     "db.example.com",
		Port:     5432,
		Username: "user",
		Password: "pass",
		Database: "testdb",
		SSLMode:  "require",
		MaxConns: 20,
		MinConns: 5,
	}

	assert.Equal(t, "db.example.com", config.Host)
	assert.Equal(t, 5432, config.Port)
	assert.Equal(t, "user", config.Username)
	assert.Equal(t, "pass", config.Password)
	assert.Equal(t, "testdb", config.Database)
	assert.Equal(t, "require", config.SSLMode)
	assert.Equal(t, 20, config.MaxConns)
	assert.Equal(t, 5, config.MinConns)
}

func TestRedisConfig_Struct(t *testing.T) {
	config := RedisConfig{
		Host:     "redis.example.com",
		Port:     6380,
		Password: "secret",
		Database: 1,
		TTL:      600,
	}

	assert.Equal(t, "redis.example.com", config.Host)
	assert.Equal(t, 6380, config.Port)
	assert.Equal(t, "secret", config.Password)
	assert.Equal(t, 1, config.Database)
	assert.Equal(t, 600, config.TTL)
}

func TestAuthConfig_Struct(t *testing.T) {
	config := AuthConfig{
		JWTSecret:     "super-secret",
		APIKeys:       []string{"key1", "key2"},
		TokenExpiry:   7200,
		RefreshExpiry: 172800,
	}

	assert.Equal(t, "super-secret", config.JWTSecret)
	assert.Equal(t, []string{"key1", "key2"}, config.APIKeys)
	assert.Equal(t, 7200, config.TokenExpiry)
	assert.Equal(t, 172800, config.RefreshExpiry)
}

func TestLoggingConfig_Struct(t *testing.T) {
	config := LoggingConfig{
		Level:  "debug",
		Format: "text",
		Output: "file",
	}

	assert.Equal(t, "debug", config.Level)
	assert.Equal(t, "text", config.Format)
	assert.Equal(t, "file", config.Output)
}

func TestConfig_CompleteStruct(t *testing.T) {
	config := Config{
		Server: ServerConfig{
			Host: "test.example.com",
			Port: 9000,
		},
		Database: DatabaseConfig{
			Host: "db.test.com",
			Port: 5433,
		},
		Redis: RedisConfig{
			Host: "redis.test.com",
			Port: 6380,
		},
		Auth: AuthConfig{
			JWTSecret: "test-jwt-secret",
		},
		Logging: LoggingConfig{
			Level: "debug",
		},
	}

	assert.Equal(t, "test.example.com", config.Server.Host)
	assert.Equal(t, 9000, config.Server.Port)
	assert.Equal(t, "db.test.com", config.Database.Host)
	assert.Equal(t, 5433, config.Database.Port)
	assert.Equal(t, "redis.test.com", config.Redis.Host)
	assert.Equal(t, 6380, config.Redis.Port)
	assert.Equal(t, "test-jwt-secret", config.Auth.JWTSecret)
	assert.Equal(t, "debug", config.Logging.Level)
}

func BenchmarkConfig_Load(b *testing.B) {
	for i := 0; i < b.N; i++ {
		viper.Reset()
		_, err := Load()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConfig_SetDefaults(b *testing.B) {
	for i := 0; i < b.N; i++ {
		viper.Reset()
		setDefaults()
	}
}