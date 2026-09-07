package commands

import (
	"fmt"
	"net/http"
)

func apiRequest(serverURL, apiKey, method, path string) (*http.Response, error) {
	req, err := http.NewRequest(method, serverURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	return response, nil
}
