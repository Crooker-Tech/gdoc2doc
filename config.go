package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// TokenData holds OAuth token information loaded from environment variable
type TokenData struct {
	Token        string   `json:"token"`
	RefreshToken string   `json:"refresh_token"`
	TokenURI     string   `json:"token_uri"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Scopes       []string `json:"scopes"`
}

// LoadGoogleDocsToken reads the OAuth token from GOOGLE-DOCS_JWT_KEY environment variable
func LoadGoogleDocsToken() (*TokenData, error) {
	tokenJSON := os.Getenv("GOOGLE-DOCS_JWT_KEY")
	if tokenJSON == "" {
		return nil, fmt.Errorf("GOOGLE-DOCS_JWT_KEY not set\n\nRun: . .\\tools\\load-key.ps1 -Service google-docs -Target JWT")
	}

	var token TokenData
	if err := json.Unmarshal([]byte(tokenJSON), &token); err != nil {
		return nil, fmt.Errorf("failed to parse GOOGLE-DOCS_JWT_KEY: %w", err)
	}

	return &token, nil
}

// GetTogetherAPIKey retrieves the Together API key from environment variable
func GetTogetherAPIKey() (string, error) {
	key := os.Getenv("TOGETHER_API_KEY")
	if key == "" {
		return "", fmt.Errorf("TOGETHER_API_KEY not set\n\nRun: . .\\tools\\load-key.ps1 -Service together -Target API")
	}
	return key, nil
}
