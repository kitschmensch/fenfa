package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	Port                    int
	Salt                    string
	DataFile                string
	DefaultExpirationPeriod int64
	ExpiredGracePeriod      int64
	Host                    string
	FailedAttemptLimit      int
	MaxZipSize              int64
	MaxZipDepth             int
	ZipDirectory            string
	TemplateIncludesPort    bool
	BinaryDirectory         string
)

func Initialize(dir string) {
	BinaryDirectory = dir
	err := godotenv.Load(BinaryDirectory + `/.env`)
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	Port, err = strconv.Atoi(os.Getenv("FENFA_PORT"))
	if err != nil {
		log.Fatal("Error, Port is missing, or Port is not an integer", err)
	}
	if Port < 1024 || Port > 65535 {
		log.Fatal("Error, invalid Port number", err)
	}

	Salt = os.Getenv("FENFA_SALT")
	Host = os.Getenv("FENFA_HOST")
	period, err := strconv.Atoi(os.Getenv("FENFA_DEFAULT_EXPIRATION_PERIOD"))
	if err != nil {
		period = 86400
	}
	DefaultExpirationPeriod = int64(period)
	ZipDirectory = BinaryDirectory + `/.fenfa`
	MaxZipDepth, err = strconv.Atoi(os.Getenv("FENFA_MAX_ZIP_DEPTH"))
	if err != nil {
		MaxZipDepth = 2
	}
	size, err := strconv.Atoi(os.Getenv("FENFA_MAX_ZIP_SIZE"))
	if err != nil {
		MaxZipSize = int64(1073741824)
	} else {
		MaxZipSize = int64(size)
	}
	DataFile = os.Getenv("FENFA_DATA_FILE")
	period2, err := strconv.Atoi(os.Getenv("FENFA_DEFAULT_EXPIRATION_PERIOD"))
	if err != nil {
		period2 = 86400
	}
	ExpiredGracePeriod = int64(period2)
	FailedAttemptLimit, err = strconv.Atoi(os.Getenv("FENFA_FAILED_ATTEMPT_LIMIT"))
	if err != nil {
		log.Fatal("Error, Failed Attempt Limit must be an int")
	}

	templatePort := os.Getenv("FENFA_TEMPLATE_INCLUDES_PORT")
	if templatePort == "" {
		TemplateIncludesPort = true
	} else {
		parsedValue, err := strconv.ParseBool(templatePort)
		if err != nil {
			log.Fatalf("Invalid value for FENFA_TEMPLATE_INCLUDES_PORT: %v", err)
		}
		TemplateIncludesPort = parsedValue
	}
}
