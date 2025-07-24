package main

import (
	"context"
	"log"
	"os"
	user_provider "vestro/internal/adaptadores/agriwin/usuario"
	agriwin_api "vestro/internal/adaptadores/agriwin_api"
	vestro_api "vestro/internal/adaptadores/vestro_api"
	servicos "vestro/internal/aplicacao/servicos"
	"vestro/internal/config"
)

func main() {
	// Carrega a configuração
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Verifica se as configurações essenciais estão presentes
	if cfg.VestroLogin == "" || cfg.VestroPassword == "" || cfg.GrailsAppURL == "" {
		log.Fatal("Essential environment variables (VESTRO_LOGIN, VESTRO_PASSWORD, GRAILS_APP_URL) are not set.")
		os.Exit(1)
	}

	// --- Composição das Dependências (Dependency Injection) ---

	// 1. Cria os adaptadores (implementações concretas das portas)
	vestroClient := vestro_api.New(cfg.VestroBaseURL, cfg.VestroLogin, cfg.VestroPassword)
	grailsNotifier := agriwin_api.New(cfg.GrailsAppURL)
	agriwinUserProvider := user_provider.New(cfg.AgriwinUsersURL)

	// 2. Cria o serviço do core, injetando os adaptadores como interfaces
	importerService := servicos.New(vestroClient, grailsNotifier, agriwinUserProvider, cfg.FetchDataSince)

	// 3. Executa o serviço
	if err := importerService.RunImport(context.Background()); err != nil {
		log.Fatalf("Job execution failed: %v", err)
		os.Exit(1) // Em um job, é importante sair com um código de erro
	}

	log.Println("Job completed successfully.")
}
