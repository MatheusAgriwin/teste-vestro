package servicos

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	"vestro/internal/aplicacao/portas"
	"vestro/internal/dto"
)

type ImporterService struct {
	apiClient    portas.VestroAPIClient
	notifier     portas.Notifier
	userProvider portas.UserProvider
	fetchSince   time.Duration
}

func New(
	apiClient portas.VestroAPIClient,
	notifier portas.Notifier,
	userProvider portas.UserProvider,
	fetchSince time.Duration,
) *ImporterService {
	return &ImporterService{
		apiClient:    apiClient,
		notifier:     notifier,
		userProvider: userProvider,
		fetchSince:   fetchSince,
	}
}

func (s *ImporterService) RunImport(ctx context.Context) error {
	log.Println("Starting Vestro data import job...")

	// 1. Buscar produtores a processar da API Agriwin
	log.Println("Fetching users to integrate from Agriwin...")
	users, err := s.userProvider.GetUsersToIntegrate(ctx)
	if err != nil {
		return fmt.Errorf("could not get users to integrate: %w", err)
	}

	if len(users) == 0 {
		log.Println("No users to integrate. Job finished.")
		return nil
	}
	log.Printf("Found %d users to process.", len(users))

	// 2. Loop para processar cada produtor individualmente
	for _, user := range users {
		log.Printf("------------------ Processing Producer ID: %d ------------------", user.ProdutorID)

		// 2.1. Autenticar na API Vestro com as credenciais do produtor atual
		log.Printf("Authenticating user '%s' with Vestro API...", user.Login)
		token, err := s.apiClient.Authenticate(ctx, user.Login, user.Senha)
		if err != nil {
			log.Printf("ERROR: Vestro authentication failed for user '%s': %v. Skipping.", user.Login, err)
			continue // Pula para o próximo produtor
		}
		log.Println("Authentication successful for this user.")

		// 2.2. Buscar todos os dados para este produtor
		lastSync := user.Data
		// Garante que não buscamos um histórico muito longo na primeira vez
		if time.Since(lastSync) > s.fetchSince {
			lastSync = time.Now().Add(-s.fetchSince)
		}

		log.Printf("Fetching data since %v", lastSync)
		userPayload, err := s.fetchAllDataForUser(ctx, token, user, lastSync)
		if err != nil {
			log.Printf("ERROR: Failed to fetch data for producer %d: %v. Skipping.", user.ProdutorID, err)
			continue
		}

		// 2.3. Enviar dados se houver algo novo
		if userPayload.IsEmpty() {
			log.Printf("No new transactional data found for producer %d.", user.ProdutorID)
			continue
		}

		log.Printf("Sending payload for producer %d to Agriwin...", user.ProdutorID)
		if err := s.notifier.Send(ctx, *userPayload); err != nil {
			log.Printf("ERROR: Failed to send data for producer %d: %v. Skipping.", user.ProdutorID, err)
			continue
		}
		log.Printf("Successfully processed producer %d.", user.ProdutorID)
	}

	log.Println("------------------ Job finished successfully ------------------")
	return nil
}

// fetchAllDataForUser busca todos os dados (mestres e transacionais) para um usuário.
func (s *ImporterService) fetchAllDataForUser(ctx context.Context, token string, user dto.UserToIntegrate, since time.Time) (*dto.IntegrationPayload, error) {
	var wg sync.WaitGroup
	errChan := make(chan error, 7) // 2 transacionais + 5 de cadastro

	payload := &dto.IntegrationPayload{
		ProdutorID: user.ProdutorID,
		FetchedAt:  time.Now(),
	}

	// O identificador na Vestro para filtrar os dados transacionais será o login.
	// A propriedade de filtro será "driver", que é um palpite comum.
	// Se o login for de um frentista, a propriedade pode ser "employee".
	vestroIdentifier := user.Login

	// --- Buscas Transacionais (com filtro de data e usuário) ---
	wg.Add(2)
	go s.fetchData(ctx, &wg, errChan, "supplies", func() (interface{}, error) {
		return s.apiClient.GetSupplies(ctx, token, since, vestroIdentifier)
	}, &payload.Supplies)
	go s.fetchData(ctx, &wg, errChan, "productSales", func() (interface{}, error) {
		return s.apiClient.GetProductSales(ctx, token, since, vestroIdentifier)
	}, &payload.ProductSales)

	// --- Buscas de Dados Mestres (sem filtro de data ou usuário específico, mas sob o token do usuário) ---
	wg.Add(5)
	go s.fetchData(ctx, &wg, errChan, "products", func() (interface{}, error) { return s.apiClient.GetProducts(ctx, token) }, &payload.Products)
	go s.fetchData(ctx, &wg, errChan, "fuelTypes", func() (interface{}, error) { return s.apiClient.GetFuelTypes(ctx, token) }, &payload.FuelTypes)
	go s.fetchData(ctx, &wg, errChan, "vehicles", func() (interface{}, error) { return s.apiClient.GetVehicles(ctx, token) }, &payload.Vehicles)
	go s.fetchData(ctx, &wg, errChan, "drivers", func() (interface{}, error) { return s.apiClient.GetDrivers(ctx, token) }, &payload.Drivers)
	go s.fetchData(ctx, &wg, errChan, "employees", func() (interface{}, error) { return s.apiClient.GetEmployees(ctx, token) }, &payload.Employees)

	wg.Wait()
	close(errChan)

	for fetchErr := range errChan {
		if fetchErr != nil {
			return nil, fetchErr
		}
	}
	return payload, nil
}

// (A função helper 'fetchData' continua a mesma)
func (s *ImporterService) fetchData(ctx context.Context, wg *sync.WaitGroup, errChan chan<- error, name string, fetchFunc func() (interface{}, error), result interface{}) {
	defer wg.Done()
	log.Printf("Fetching %s...", name)
	data, err := fetchFunc()
	if err != nil {
		errChan <- fmt.Errorf("failed to fetch %s: %w", name, err)
		return
	}

	switch r := result.(type) {
	case *[]dto.Supply:
		*r = data.([]dto.Supply)
	case *[]dto.ProductSale:
		*r = data.([]dto.ProductSale)
	case *[]dto.Product:
		*r = data.([]dto.Product)
	case *[]dto.FuelType:
		*r = data.([]dto.FuelType)
	case *[]dto.Vehicle:
		*r = data.([]dto.Vehicle)
	case *[]dto.Driver:
		*r = data.([]dto.Driver)
	case *[]dto.Employee:
		*r = data.([]dto.Employee)
	}

	log.Printf("Successfully fetched %s.", name)
}
