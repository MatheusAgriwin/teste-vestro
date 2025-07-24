package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	portas "vestro/internal/aplicacao/portas"
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

// RunImport executa o novo processo iterativo.
func (s *ImporterService) RunImport(ctx context.Context) error {
	log.Println("Starting Vestro data import job...")

	// 1. Buscar usuários a processar
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

	// 2. Autenticar na API Vestro (uma vez)
	log.Println("Authenticating with Vestro API...")
	token, err := s.apiClient.Authenticate(ctx)
	if err != nil {
		return fmt.Errorf("vestro authentication failed: %w", err)
	}
	log.Println("Authentication successful.")

	// Carrega dados que são comuns a todos (produtos, tipos de combustível, etc)
	commonPayload, err := s.fetchCommonData(ctx, token)
	if err != nil {
		log.Printf("Warning: failed to fetch some common data: %v", err)
	}

	// 3. Loop para processar cada usuário
	for _, user := range users {
		log.Printf("Processing user: %s (Vestro ID: %s)", user.UserUUID, user.VestroIdentifier)

		lastSync := user.LastIntegration
		// Se a data for muito antiga, usa o padrão do job para não buscar dados demais
		if time.Since(lastSync) > s.fetchSince {
			lastSync = time.Now().Add(-s.fetchSince)
		}

		userPayload, err := s.fetchDataForUser(ctx, token, user, lastSync)
		if err != nil {
			log.Printf("ERROR: Failed to fetch data for user %s: %v. Skipping.", user.UserUUID, err)
			continue // Pula para o próximo usuário
		}

		// Combina os dados comuns com os dados do usuário
		userPayload.Products = commonPayload.Products
		userPayload.FuelTypes = commonPayload.FuelTypes
		userPayload.Vehicles = commonPayload.Vehicles
		userPayload.Drivers = commonPayload.Drivers
		userPayload.Employees = commonPayload.Employees

		if userPayload.IsEmpty() {
			log.Printf("No new data found for user %s.", user.UserUUID)
			continue
		}

		log.Printf("Sending payload for user %s to Agriwin...", user.UserUUID)
		if err := s.notifier.Send(ctx, *userPayload); err != nil {
			log.Printf("ERROR: Failed to send data for user %s: %v. Skipping.", user.UserUUID, err)
			continue
		}
		log.Printf("Successfully processed user %s.", user.UserUUID)
	}

	log.Println("Job finished successfully.")
	return nil
}

// fetchCommonData busca dados que não são específicos do usuário.
func (s *ImporterService) fetchCommonData(ctx context.Context, token string) (*dto.IntegrationPayload, error) {
	var wg sync.WaitGroup
	errChan := make(chan error, 5)
	payload := &dto.IntegrationPayload{}

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

// fetchDataForUser busca dados que SÃO específicos do usuário.
func (s *ImporterService) fetchDataForUser(ctx context.Context, token string, user dto.UserToIntegrate, since time.Time) (*dto.IntegrationPayload, error) {
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	payload := &dto.IntegrationPayload{
		UserUUID:  user.UserUUID,
		FetchedAt: time.Now(),
	}

	wg.Add(2)
	go s.fetchData(ctx, &wg, errChan, "supplies", func() (interface{}, error) {
		return s.apiClient.GetSupplies(ctx, token, since, user.VestroIdentifier)
	}, &payload.Supplies)

	go s.fetchData(ctx, &wg, errChan, "productSales", func() (interface{}, error) {
		return s.apiClient.GetProductSales(ctx, token, since, user.VestroIdentifier)
	}, &payload.ProductSales)

	wg.Wait()
	close(errChan)

	// Verifica se houve erros durante a busca
	for fetchErr := range errChan {
		if fetchErr != nil {
			// Para a busca de um usuário, qualquer erro é fatal para aquele usuário.
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
