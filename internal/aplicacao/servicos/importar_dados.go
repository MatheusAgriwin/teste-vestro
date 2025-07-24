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
	apiClient  portas.VestroAPIClient
	notifier   portas.Notifier
	fetchSince time.Duration
}

func New(
	apiClient portas.VestroAPIClient,
	notifier portas.Notifier,
	fetchSince time.Duration,
) *ImporterService {
	return &ImporterService{
		apiClient:  apiClient,
		notifier:   notifier,
		fetchSince: fetchSince,
	}
}

// RunImport executa o processo completo de importação e notificação.
func (s *ImporterService) RunImport(ctx context.Context) error {
	log.Println("Starting Vestro data import job...")

	// 1. Autenticar na API Vestro
	log.Println("Authenticating with Vestro API...")
	token, err := s.apiClient.Authenticate(ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	log.Println("Authentication successful.")

	// 2. Buscar todos os dados em paralelo
	var wg sync.WaitGroup
	errChan := make(chan error, 7) // Um buffer para cada goroutine

	payload := dto.IntegrationPayload{
		FetchedAt: time.Now(),
	}

	since := time.Now().Add(-s.fetchSince)
	log.Printf("Fetching data since %v", since)

	// Funções que dependem de data
	wg.Add(2)
	go s.fetchData(ctx, &wg, errChan, "supplies", func() (interface{}, error) { return s.apiClient.GetSupplies(ctx, token, since) }, &payload.Supplies)
	go s.fetchData(ctx, &wg, errChan, "productSales", func() (interface{}, error) { return s.apiClient.GetProductSales(ctx, token, since) }, &payload.ProductSales)

	// Funções de cadastro (não dependem de data)
	wg.Add(5)
	go s.fetchData(ctx, &wg, errChan, "products", func() (interface{}, error) { return s.apiClient.GetProducts(ctx, token) }, &payload.Products)
	go s.fetchData(ctx, &wg, errChan, "fuelTypes", func() (interface{}, error) { return s.apiClient.GetFuelTypes(ctx, token) }, &payload.FuelTypes)
	go s.fetchData(ctx, &wg, errChan, "vehicles", func() (interface{}, error) { return s.apiClient.GetVehicles(ctx, token) }, &payload.Vehicles)
	go s.fetchData(ctx, &wg, errChan, "drivers", func() (interface{}, error) { return s.apiClient.GetDrivers(ctx, token) }, &payload.Drivers)
	go s.fetchData(ctx, &wg, errChan, "employees", func() (interface{}, error) { return s.apiClient.GetEmployees(ctx, token) }, &payload.Employees)

	wg.Wait()
	close(errChan)

	// Verificar se houve erros durante a busca
	for fetchErr := range errChan {
		if fetchErr != nil {
			// Pode-se decidir continuar mesmo com erro em um endpoint, ou parar tudo.
			// Aqui, vamos apenas logar o erro e continuar.
			log.Printf("Error during data fetching: %v", fetchErr)
		}
	}

	// 3. Enviar os dados para a aplicação Grails se houver algo novo
	if payload.IsEmpty() {
		log.Println("No new data to send. Job finished.")
		return nil
	}

	log.Printf("Sending %d supplies, %d product sales, %d products, etc. to Grails application...",
		len(payload.Supplies), len(payload.ProductSales), len(payload.Products))

	if err := s.notifier.Send(ctx, payload); err != nil {
		return fmt.Errorf("failed to send data to notifier: %w", err)
	}

	log.Println("Data sent successfully to Grails application. Job finished.")
	return nil
}

// fetchData é um helper para executar as chamadas em paralelo.
func (s *ImporterService) fetchData(ctx context.Context, wg *sync.WaitGroup, errChan chan<- error, name string, fetchFunc func() (interface{}, error), result interface{}) {
	defer wg.Done()
	log.Printf("Fetching %s...", name)
	data, err := fetchFunc()
	if err != nil {
		errChan <- fmt.Errorf("failed to fetch %s: %w", name, err)
		return
	}

	// Usando reflection para atribuir o resultado ao campo correto do payload
	// Isso é complexo, então para simplificar vamos fazer a atribuição direta após a chamada
	// A estrutura aqui é um exemplo de como seria. O código atualizado está no RunImport.
	// Por simplicidade, faremos a atribuição nos callers.

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
