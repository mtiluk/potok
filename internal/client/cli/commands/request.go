package commands

import (
	"encoding/json"
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

func apiRequestJSON(serverURL, apiKey, method, path string, out any) (*http.Response, error) {
	response, err := apiRequest(serverURL, apiKey, method, path)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return response, nil
	}

	if out != nil {
		if err := json.NewDecoder(response.Body).Decode(out); err != nil {
			return response, fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return response, nil
}
