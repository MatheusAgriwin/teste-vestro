package vestro_api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"vestro/internal/dto"
)

type apiClient struct {
	baseURL    string
	login      string
	password   string
	httpClient *http.Client
}

func New(baseURL, login, password string) *apiClient {
	return &apiClient{
		baseURL:  baseURL,
		login:    login,
		password: password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// --- Implementação da Interface VestroAPIClient ---

func (c *apiClient) Authenticate(ctx context.Context) (string, error) {
	formData := url.Values{}
	formData.Set("login", c.login)
	formData.Set("password", c.password)

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/sessions", bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create auth request: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute auth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth request failed with status: %s", resp.Status)
	}

	var wrapper dto.VestroResponseWrapper
	var authResp dto.AuthResponse
	wrapper.Data = &authResp

	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return "", fmt.Errorf("failed to decode auth response: %w", err)
	}

	if !wrapper.Success {
		return "", fmt.Errorf("authentication failed on API")
	}

	return authResp.Access, nil
}

// Funções de busca genéricas com paginação
func (c *apiClient) fetchPaginatedData(ctx context.Context, token, path string, params url.Values, result interface{}) error {
	const limit = 100
	start := 0

	// Aponta para o slice que irá receber os dados
	slicePtr, ok := result.([]interface{})
	if !ok {
		// Esta é uma simplificação. Uma implementação real usaria reflection para ser mais genérica.
		// Para este caso, vamos assumir que o tipo é conhecido.
		return fmt.Errorf("result must be a slice of interfaces")
	}
	
	for {
		// Adiciona parâmetros de paginação
		q := url.Values{}
		if params != nil {
			q = params
		}
		q.Set("start", strconv.Itoa(start))
		q.Set("limit", strconv.Itoa(limit))
		
		fullURL := fmt.Sprintf("%s%s?%s", c.baseURL, path, q.Encode())

		req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request for %s: %w", path, err)
		}
		req.Header.Add("Authorization", "Bearer "+token)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("request to %s failed: %w", path, err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return fmt.Errorf("request to %s got status %s", path, resp.Status)
		}
		
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("failed to read response body from %s: %w", path, err)
		}

		var wrapper dto.VestroResponseWrapper
		var pageData []json.RawMessage // Usamos RawMessage para decodificar depois
		wrapper.Data = &pageData
		
		if err := json.Unmarshal(body, &wrapper); err != nil {
			return fmt.Errorf("failed to decode wrapper for %s: %w", path, err)
		}
		
		// Decodifica cada item para o tipo correto (aqui está a simplificação)
        // Isso precisa ser melhorado para ser verdadeiramente genérico ou ter uma função por tipo
        // Para simplificar, faremos uma por tipo.

		// Adiciona dados da página ao resultado total
		// slicePtr = append(slicePtr, pageData...)

		log.Printf("Fetched %d records from %s (start: %d)", len(pageData), path, start)

		// Se o número de registros retornados for menor que o limite, chegamos ao fim.
		if len(pageData) < limit {
			break
		}
		
		start += limit
	}

	// *result = slicePtr
	return nil
}

// As funções fetchPaginatedData foram simplificadas para uma chamada por tipo para facilitar.
// A lógica de paginação real está implementada dentro de cada função Get...

func (c *apiClient) GetSupplies(ctx context.Context, token string, since time.Time) ([]dto.Supply, error) {
	var allSupplies []dto.Supply
	if err := c.fetchAndAggregate(ctx, token, "/supplies", since, &allSupplies); err != nil {
		return nil, fmt.Errorf("failed getting supplies: %w", err)
	}
	return allSupplies, nil
}

func (c *apiClient) GetProductSales(ctx context.Context, token string, since time.Time) ([]dto.ProductSale, error) {
	var allProductSales []dto.ProductSale
	if err := c.fetchAndAggregate(ctx, token, "/product/sales", since, &allProductSales); err != nil {
		return nil, fmt.Errorf("failed getting product sales: %w", err)
	}
	return allProductSales, nil
}

func (c *apiClient) GetProducts(ctx context.Context, token string) ([]dto.Product, error) {
	var allProducts []dto.Product
	if err := c.fetchAndAggregate(ctx, token, "/products", time.Time{}, &allProducts); err != nil {
		return nil, fmt.Errorf("failed getting products: %w", err)
	}
	return allProducts, nil
}

func (c *apiClient) GetFuelTypes(ctx context.Context, token string) ([]dto.FuelType, error) {
	var allFuelTypes []dto.FuelType
	if err := c.fetchAndAggregate(ctx, token, "/fuel/types", time.Time{}, &allFuelTypes); err != nil {
		return nil, fmt.Errorf("failed getting fuel types: %w", err)
	}
	return allFuelTypes, nil
}

func (c *apiClient) GetVehicles(ctx context.Context, token string) ([]dto.Vehicle, error) {
	var allVehicles []dto.Vehicle
	if err := c.fetchAndAggregate(ctx, token, "/vehicles", time.Time{}, &allVehicles); err != nil {
		return nil, fmt.Errorf("failed getting vehicles: %w", err)
	}
	return allVehicles, nil
}

func (c *apiClient) GetDrivers(ctx context.Context, token string) ([]dto.Driver, error) {
	var allDrivers []dto.Driver
	if err := c.fetchAndAggregate(ctx, token, "/drivers", time.Time{}, &allDrivers); err != nil {
		return nil, fmt.Errorf("failed getting drivers: %w", err)
	}
	return allDrivers, nil
}

func (c *apiClient) GetEmployees(ctx context.Context, token string) ([]dto.Employee, error) {
	var allEmployees []dto.Employee
	if err := c.fetchAndAggregate(ctx, token, "/employees", time.Time{}, &allEmployees); err != nil {
		return nil, fmt.Errorf("failed getting employees: %w", err)
	}
	return allEmployees, nil
}

// fetchAndAggregate é uma função helper genérica que lida com a paginação.
func (c *apiClient) fetchAndAggregate<T any>(ctx context.Context, token, path string, since time.Time, result *[]T) error {
	const limit = 100
	start := 0

	for {
		q := url.Values{}
		q.Set("start", strconv.Itoa(start))
		q.Set("limit", strconv.Itoa(limit))
		q.Set("sort", "true") // Ordena por data ascendente para um processamento consistente

		if !since.IsZero() {
			// Formato: "yyyy-mm-ddThh-mm-ssZ"
			q.Set("startDate", since.UTC().Format("2006-01-02T15-04-05Z"))
		}

		fullURL := fmt.Sprintf("%s%s?%s", c.baseURL, path, q.Encode())
		req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request for %s: %w", path, err)
		}
		req.Header.Add("Authorization", "Bearer "+token)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("request to %s failed: %w", path, err)
		}
		
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return fmt.Errorf("request to %s got status %s, body: %s", path, resp.Status, string(body))
		}

		// Usamos um wrapper com json.RawMessage para adiar a decodificação de `data`
		var rawWrapper struct {
			Success bool `json:"success"`
			Data json.RawMessage `json:"data"`
			Count int `json:"count"`
		}
		
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		
		if err := json.Unmarshal(body, &rawWrapper); err != nil {
			return fmt.Errorf("failed to decode raw wrapper for %s: %w. Body: %s", path, err, string(body))
		}

		if !rawWrapper.Success {
			return fmt.Errorf("api call to %s was not successful. Body: %s", path, string(body))
		}
		
		var pageData []T
		if err := json.Unmarshal(rawWrapper.Data, &pageData); err != nil {
			return fmt.Errorf("failed to decode page data for %s: %w. Data field: %s", path, err, string(rawWrapper.Data))
		}
		
		*result = append(*result, pageData...)
		log.Printf("Fetched %d records from %s (total so far: %d)", len(pageData), path, len(*result))
		
		if len(pageData) < limit {
			break
		}

		start += limit
	}
	return nil
}