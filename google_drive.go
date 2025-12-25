package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// ExportFormat defines a supported export format
type ExportFormat struct {
	MimeType  string
	Extension string
}

// Supported export formats for Google Docs
var ExportFormats = map[string]ExportFormat{
	"pdf":      {MimeType: "application/pdf", Extension: ".pdf"},
	"docx":     {MimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document", Extension: ".docx"},
	"odt":      {MimeType: "application/vnd.oasis.opendocument.text", Extension: ".odt"},
	"rtf":      {MimeType: "application/rtf", Extension: ".rtf"},
	"txt":      {MimeType: "text/plain", Extension: ".txt"},
	"html":     {MimeType: "text/html", Extension: ".html"},
	"epub":     {MimeType: "application/epub+zip", Extension: ".epub"},
	"md":       {MimeType: "text/markdown", Extension: ".md"},
	"markdown": {MimeType: "text/markdown", Extension: ".md"},
}


// GoogleDocument represents a Google Doc with its metadata
type GoogleDocument struct {
	ID          string
	Name        string
	Description string
	ModifiedAt  string
	CreatedAt   string
}

// DriveClient wraps Google Drive API operations
type DriveClient struct {
	service *drive.Service
}

// NewDriveClient creates a new Drive client from token data
func NewDriveClient(token *TokenData) (*DriveClient, error) {
	ctx := context.Background()

	config := &oauth2.Config{
		ClientID:     token.ClientID,
		ClientSecret: token.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       token.Scopes,
	}

	oauthToken := &oauth2.Token{
		AccessToken:  token.Token,
		RefreshToken: token.RefreshToken,
		TokenType:    "Bearer",
	}

	client := config.Client(ctx, oauthToken)

	service, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Drive service: %w", err)
	}

	return &DriveClient{service: service}, nil
}

// ListGoogleDocs returns all Google Docs in the user's Drive
func (client *DriveClient) ListGoogleDocs() ([]GoogleDocument, error) {
	var documents []GoogleDocument

	pageToken := ""
	for {
		query := client.service.Files.List().
			Q("mimeType='application/vnd.google-apps.document' and trashed=false").
			Fields("nextPageToken, files(id, name, description, modifiedTime, createdTime)").
			PageSize(100)

		if pageToken != "" {
			query = query.PageToken(pageToken)
		}

		response, err := query.Do()
		if err != nil {
			return nil, fmt.Errorf("failed to list files: %w", err)
		}

		for _, file := range response.Files {
			documents = append(documents, GoogleDocument{
				ID:          file.Id,
				Name:        file.Name,
				Description: file.Description,
				ModifiedAt:  file.ModifiedTime,
				CreatedAt:   file.CreatedTime,
			})
		}

		pageToken = response.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return documents, nil
}

// ExportDocument exports a Google Doc to the specified format
func (client *DriveClient) ExportDocument(documentID string, outputPath string, format ExportFormat) error {
	response, err := client.service.Files.Export(documentID, format.MimeType).Download()
	if err != nil {
		return fmt.Errorf("failed to export document: %w", err)
	}
	defer response.Body.Close()

	// Ensure output directory exists
	outputDirectory := filepath.Dir(outputPath)
	if outputDirectory != "." && outputDirectory != "" {
		if err := os.MkdirAll(outputDirectory, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	bytesWritten, err := io.Copy(outputFile, response.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Exported %d bytes to %s\n", bytesWritten, outputPath)
	return nil
}

// SanitizeFilename removes characters that are invalid in filenames
func SanitizeFilename(name string) string {
	invalidChars := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}
