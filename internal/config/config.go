package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

// Constants for environment variable keys
const (
	EnvPort                    = "FENFA_PORT"
	EnvDataFile                = "FENFA_DATA_FILE"
	EnvDefaultExpirationPeriod = "FENFA_DEFAULT_EXPIRATION_PERIOD"
	EnvExpiredGracePeriod      = "FENFA_EXPIRED_GRACE_PERIOD"
	EnvHost                    = "FENFA_HOST"
	EnvFailedAttemptLimit      = "FENFA_FAILED_ATTEMPT_LIMIT"
	EnvMaxZipSize              = "FENFA_MAX_ZIP_SIZE"
	EnvMaxZipDepth             = "FENFA_MAX_ZIP_DEPTH"
	EnvTemplateIncludesPort    = "FENFA_TEMPLATE_INCLUDES_PORT"
	EnvRateLimit               = "FENFA_RATE_LIMIT"
)

// Default values
const (
	DefaultPort               = 8080
	DefaultExpirationPeriod   = 86400      // 24 hours
	DefaultMaxZipSize         = 1073741824 // 1 GB
	DefaultMaxZipDepth        = 2
	DefaultFailedAttemptLimit = 5
	DefaultRateLimit          = 30
)

// Global configuration variables
var (
	Port                 int
	Salt                 string
	DataFile             string
	ExpirationPeriod     int64
	ExpiredGracePeriod   int64
	Host                 string
	FailedAttemptLimit   int
	MaxZipSize           int64
	MaxZipDepth          int
	ZipDirectory         string
	TemplateIncludesPort bool
	BinaryDirectory      string
	RateLimit            int
)

// Initialize loads configuration from the environment
func Initialize(dir string) {
	BinaryDirectory = dir
	err := godotenv.Load(filepath.Join(BinaryDirectory, ".env"))
	if err != nil {
		log.Fatal("Error loading .env file: ", err)
	}

	Port = getEnvAsInt(EnvPort, DefaultPort)
	Host = os.Getenv(EnvHost)
	ExpirationPeriod = getEnvAsInt64(EnvDefaultExpirationPeriod, DefaultExpirationPeriod)
	ExpiredGracePeriod = getEnvAsInt64(EnvExpiredGracePeriod, DefaultExpirationPeriod)
	FailedAttemptLimit = getEnvAsInt(EnvFailedAttemptLimit, DefaultFailedAttemptLimit)
	MaxZipDepth = getEnvAsInt(EnvMaxZipDepth, DefaultMaxZipDepth)
	MaxZipSize = getEnvAsInt64(EnvMaxZipSize, DefaultMaxZipSize)
	RateLimit = getEnvAsInt(EnvRateLimit, DefaultRateLimit)
	TemplateIncludesPort = getEnvAsBool(EnvTemplateIncludesPort, true)

	// DataFile and ZipDirectory require additional setup
	DataFile = os.Getenv(EnvDataFile)
	ZipDirectory = filepath.Join(BinaryDirectory, ".fenfa")
}

// Helper to get environment variables as an integer with a default value
func getEnvAsInt(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		log.Fatalf("Invalid integer value for %s: %v", key, err)
	}
	return val
}

// Helper to get environment variables as an int64 with a default value
func getEnvAsInt64(key string, defaultVal int64) int64 {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid int64 value for %s: %v", key, err)
	}
	return val
}

// Helper to get environment variables as a boolean with a default value
func getEnvAsBool(key string, defaultVal bool) bool {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		log.Fatalf("Invalid boolean value for %s: %v", key, err)
	}
	return val
}
