package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config armazena todas as configurações da aplicação.
type Config struct {
	VestroBaseURL   string
	GrailsAppURL    string
	AgriwinUsersURL string
	FetchDataSince  time.Duration
}

// Load carrega as configurações das variáveis de ambiente.
func Load() (*Config, error) {
	_ = godotenv.Load()

	fetchHours, err := strconv.Atoi(getEnv("FETCH_DATA_SINCE_HOURS", "24")) // Aumentado para 24h como padrão
	if err != nil {
		log.Printf("Invalid FETCH_DATA_SINCE_HOURS, using default 24h. Error: %v", err)
		fetchHours = 24
	}

	return &Config{
		VestroBaseURL:   getEnv("VESTRO_API_URL", ""),
		GrailsAppURL:    getEnv("GRAILS_APP_URL", ""),
		AgriwinUsersURL: getEnv("AGRIWIN_USERS_URL", ""),
		FetchDataSince:  time.Duration(fetchHours) * time.Hour,
	}, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Printf("Environment variable %s not set, using fallback: '%s'", key, fallback)
	return fallback
}
