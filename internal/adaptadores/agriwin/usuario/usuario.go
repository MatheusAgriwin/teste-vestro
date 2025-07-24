package usuario

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"vestro/internal/dto"
)

type userProvider struct {
	usersURL   string
	httpClient *http.Client
}

func New(usersURL string) *userProvider {
	return &userProvider{
		usersURL: usersURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *userProvider) GetUsersToIntegrate(ctx context.Context) ([]dto.UserToIntegrate, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.usersURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for agriwin users: %w", err)
	}
	// Se precisar de autenticação, adicione o header aqui
	// req.Header.Set("Authorization", "Bearer your_token")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get users from agriwin: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("agriwin users endpoint responded with status: %s", resp.Status)
	}

	var users []dto.UserToIntegrate
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to decode agriwin users response: %w", err)
	}

	return users, nil
}
