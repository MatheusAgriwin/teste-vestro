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
	VestroBaseURL  string
	VestroLogin    string
	VestroPassword string
	GrailsAppURL   string
	FetchDataSince time.Duration
}

// Load carrega as configurações das variáveis de ambiente.
// O arquivo .env é usado apenas para desenvolvimento local.
func Load() (*Config, error) {
	// Ignora erro se o arquivo .env não for encontrado (comum em produção)
	_ = godotenv.Load()

	fetchHours, err := strconv.Atoi(getEnv("FETCH_DATA_SINCE_HOURS", "1"))
	if err != nil {
		log.Printf("Invalid FETCH_DATA_SINCE_HOURS, using default 1h. Error: %v", err)
		fetchHours = 1
	}

	return &Config{
		VestroBaseURL:  getEnv("VESTRO_API_URL", "http://dev.api.vestroeletronicos.com.br:3001"),
		VestroLogin:    getEnv("VESTRO_LOGIN", ""),
		VestroPassword: getEnv("VESTRO_PASSWORD", ""),
		GrailsAppURL:   getEnv("GRAILS_APP_URL", ""),
		FetchDataSince: time.Duration(fetchHours) * time.Hour,
	}, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Printf("Environment variable %s not set, using fallback: '%s'", key, fallback)
	return fallback
}
