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

	var wrapper struct {
		Success bool             `json:"success"`
		Data    dto.AuthResponse `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return "", fmt.Errorf("failed to decode auth response: %w", err)
	}

	if !wrapper.Success {
		return "", fmt.Errorf("authentication failed on API")
	}

	return wrapper.Data.Access, nil
}

// As funções abaixo agora chamam a *função* genérica fetchAndAggregate,
// passando as dependências do apiClient.
func (c *apiClient) GetSupplies(ctx context.Context, token string, since time.Time, userIdentifier string) ([]dto.Supply, error) {
	// A propriedade de filtro 'driver' é um palpite. Pode ser 'employee' ou outra.
	return fetchAndAggregate[dto.Supply](ctx, c.httpClient, c.baseURL, token, "/supplies", since, "driver", userIdentifier)
}

func (c *apiClient) GetProductSales(ctx context.Context, token string, since time.Time, userIdentifier string) ([]dto.ProductSale, error) {
	return fetchAndAggregate[dto.ProductSale](ctx, c.httpClient, c.baseURL, token, "/product/sales", since, "driver", userIdentifier)
}

func (c *apiClient) GetProducts(ctx context.Context, token string) ([]dto.Product, error) {
	return fetchAndAggregate[dto.Product](ctx, c.httpClient, c.baseURL, token, "/products", time.Time{}, "", "")
}

func (c *apiClient) GetFuelTypes(ctx context.Context, token string) ([]dto.FuelType, error) {
	return fetchAndAggregate[dto.FuelType](ctx, c.httpClient, c.baseURL, token, "/fuel/types", time.Time{}, "", "")
}

func (c *apiClient) GetVehicles(ctx context.Context, token string) ([]dto.Vehicle, error) {
	return fetchAndAggregate[dto.Vehicle](ctx, c.httpClient, c.baseURL, token, "/vehicles", time.Time{}, "", "")
}

func (c *apiClient) GetDrivers(ctx context.Context, token string) ([]dto.Driver, error) {
	return fetchAndAggregate[dto.Driver](ctx, c.httpClient, c.baseURL, token, "/drivers", time.Time{}, "", "")
}

func (c *apiClient) GetEmployees(ctx context.Context, token string) ([]dto.Employee, error) {
	return fetchAndAggregate[dto.Employee](ctx, c.httpClient, c.baseURL, token, "/employees", time.Time{}, "", "")
}

// fetchAndAggregate é agora uma FUNÇÃO genérica, não um método.
// Ela recebe httpClient e baseURL como parâmetros.
func fetchAndAggregate[T any](ctx context.Context, httpClient *http.Client, baseURL, token, path string, since time.Time, filterProperty, filterValue string) ([]T, error) {
	var allResults []T
	const limit = 100
	start := 0

	for {
		q := url.Values{}
		q.Set("start", strconv.Itoa(start))
		q.Set("limit", strconv.Itoa(limit))
		q.Set("sort", "true")

		if !since.IsZero() {
			q.Set("startDate", since.UTC().Format("2006-01-02T15-04-05Z"))
		}

		if filterProperty != "" && filterValue != "" {
			q.Set("property", filterProperty)
			q.Set("search", filterValue)
		}
		fullURL := fmt.Sprintf("%s%s?%s", baseURL, path, q.Encode())
		req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request for %s: %w", path, err)
		}
		req.Header.Add("Authorization", "Bearer "+token)

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request to %s failed: %w", path, err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("request to %s got status %s, body: %s", path, resp.Status, string(body))
		}

		var wrapper struct {
			Success bool              `json:"success"`
			Data    []json.RawMessage `json:"data"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode wrapper for %s: %w", path, err)
		}
		resp.Body.Close()

		if !wrapper.Success {
			return nil, fmt.Errorf("api call to %s was not successful", path)
		}

		// Decodifica cada item da página para o tipo genérico T
		for _, raw := range wrapper.Data {
			var item T
			if err := json.Unmarshal(raw, &item); err != nil {
				// Apenas loga o erro e continua, para não parar o job por um único registro malformado
				log.Printf("Warning: failed to unmarshal item from %s: %v", path, err)
				continue
			}
			allResults = append(allResults, item)
		}

		log.Printf("Fetched %d records from %s (total so far: %d)", len(wrapper.Data), path, len(allResults))

		if len(wrapper.Data) < limit {
			break
		}

		start += limit
	}
	return allResults, nil
}
