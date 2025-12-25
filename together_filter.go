package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// TogetherAI free model - Apriel 1.6 15B Thinker (free usage tier)
const TogetherModel = "ServiceNow-AI/Apriel-1.6-15b-Thinker"

// TogetherRequest represents a chat completion request
type TogetherRequest struct {
	Model    string           `json:"model"`
	Messages []TogetherMessage `json:"messages"`
}

// TogetherMessage represents a single message in the conversation
type TogetherMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// TogetherResponse represents the API response
type TogetherResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// DocumentFilter uses Together AI to filter documents based on natural language
type DocumentFilter struct {
	apiKey string
}

// NewDocumentFilter creates a new filter with the given API key
func NewDocumentFilter(apiKey string) *DocumentFilter {
	return &DocumentFilter{apiKey: apiKey}
}

// FilterDocuments sends document list to Together AI and returns matching document names
func (filter *DocumentFilter) FilterDocuments(documents []GoogleDocument, query string) ([]string, error) {
	// Build the document list for the prompt
	var documentList strings.Builder
	for index, document := range documents {
		documentList.WriteString(fmt.Sprintf("%d. %s", index+1, document.Name))
		if document.Description != "" {
			documentList.WriteString(fmt.Sprintf(" - %s", document.Description))
		}
		documentList.WriteString("\n")
	}

	prompt := fmt.Sprintf(`You are a document filter assistant. Given a list of document names and a search query, return ONLY the names of documents that match the query.

Which documents from this list:
%s
Match the prompt: %s

Rules:
1. Return ONLY the document names that match, one per line
2. Use EXACT document names from the list
3. If no documents match, return "NONE"
4. Do not explain or add any other text

Matching documents:`, documentList.String(), query)

	request := TogetherRequest{
		Model: TogetherModel,
		Messages: []TogetherMessage{
			{Role: "user", Content: prompt},
		},
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpRequest, err := http.NewRequest("POST", "https://api.together.xyz/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpRequest.Header.Set("Authorization", "Bearer "+filter.apiKey)
	httpRequest.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	httpResponse, err := client.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResponse.Body.Close()

	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response TogetherResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI model")
	}

	// Parse the response to extract matching document names
	content := strings.TrimSpace(response.Choices[0].Message.Content)

	if content == "NONE" || content == "" {
		return []string{}, nil
	}

	// Split by newlines and clean up
	lines := strings.Split(content, "\n")
	var matchingNames []string

	for _, line := range lines {
		cleaned := strings.TrimSpace(line)
		// Remove any bullet points or numbering
		cleaned = strings.TrimPrefix(cleaned, "- ")
		cleaned = strings.TrimPrefix(cleaned, "* ")
		// Remove numbered prefixes like "1. "
		if len(cleaned) > 3 && cleaned[1] == '.' && cleaned[2] == ' ' {
			cleaned = cleaned[3:]
		}
		if len(cleaned) > 4 && cleaned[2] == '.' && cleaned[3] == ' ' {
			cleaned = cleaned[4:]
		}
		cleaned = strings.TrimSpace(cleaned)

		if cleaned != "" && cleaned != "NONE" {
			matchingNames = append(matchingNames, cleaned)
		}
	}

	return matchingNames, nil
}

// FindMatchingDocuments returns full document info for matching names
func FindMatchingDocuments(documents []GoogleDocument, matchingNames []string) []GoogleDocument {
	var matches []GoogleDocument

	for _, document := range documents {
		for _, name := range matchingNames {
			// Flexible matching: exact match or contains
			if strings.EqualFold(document.Name, name) || strings.Contains(strings.ToLower(document.Name), strings.ToLower(name)) {
				matches = append(matches, document)
				break
			}
		}
	}

	return matches
}
