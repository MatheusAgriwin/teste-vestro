package grails_notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"vestro/internal/dto"
)

type notifier struct {
	grailsURL  string
	httpClient *http.Client
}

func New(grailsURL string) *notifier {
	return &notifier{
		grailsURL: grailsURL,
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (n *notifier) Send(ctx context.Context, payload dto.IntegrationPayload) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload for grails: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", n.grailsURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request for grails: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Se a API Grails precisar de um token, adicione aqui:
	// req.Header.Set("Authorization", "Bearer your_grails_api_token")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send data to grails: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("grails application responded with non-success status: %s", resp.Status)
	}

	return nil
}
