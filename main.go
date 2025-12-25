package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	// Parse command line flags
	outputDirectory := flag.String("output", "downloads", "Output directory for exported files")
	listOnly := flag.Bool("list", false, "List all documents without filtering")
	exportType := flag.String("t", "pdf", "Export format: pdf, docx, odt, rtf, txt, html, epub, md")
	flag.StringVar(exportType, "type", "pdf", "Export format: pdf, docx, odt, rtf, txt, html, epub, md")
	flag.Parse()

	// Get the query from remaining args
	query := strings.Join(flag.Args(), " ")

	if query == "" && !*listOnly {
		fmt.Println("gdoc2doc - Export Google Docs using natural language queries")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  gdoc2doc [flags] <query>")
		fmt.Println()
		fmt.Println("Flags:")
		fmt.Println("  -output <dir>   Output directory (default: downloads/)")
		fmt.Println("  -t, -type <fmt> Export format (default: pdf)")
		fmt.Println("                  Formats: pdf, docx, odt, rtf, txt, html, epub, md")
		fmt.Println("  -list           List all documents without filtering")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  gdoc2doc \"meeting notes\"")
		fmt.Println("  gdoc2doc -t md \"project proposal\"")
		fmt.Println("  gdoc2doc -type docx -output ./exports \"report\"")
		fmt.Println("  gdoc2doc -list")
		fmt.Println()
		fmt.Println("Setup:")
		fmt.Println("  Load API keys before running:")
		fmt.Println("    . .\\tools\\load-key.ps1 -Service together -Target API")
		fmt.Println("    . .\\tools\\load-key.ps1 -Service google-docs -Target JWT")
		os.Exit(0)
	}

	// Validate export format
	format, validFormat := ExportFormats[strings.ToLower(*exportType)]
	if !validFormat {
		fmt.Fprintf(os.Stderr, "Error: unsupported format '%s'\n", *exportType)
		fmt.Fprintf(os.Stderr, "Supported formats: pdf, docx, odt, rtf, txt, html, epub, md\n")
		os.Exit(1)
	}

	// Load Google OAuth token from environment variable
	token, err := LoadGoogleDocsToken()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create Drive client
	driveClient, err := NewDriveClient(token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// List all Google Docs
	fmt.Println("Fetching documents from Google Drive...")
	documents, err := driveClient.ListGoogleDocs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(documents) == 0 {
		fmt.Println("No Google Docs found in your Drive.")
		os.Exit(0)
	}

	fmt.Printf("Found %d documents.\n\n", len(documents))

	// If list-only mode, just print documents and exit
	if *listOnly {
		for index, document := range documents {
			fmt.Printf("%d. %s\n", index+1, document.Name)
			if document.Description != "" {
				fmt.Printf("   Description: %s\n", document.Description)
			}
			fmt.Printf("   Modified: %s\n", document.ModifiedAt)
		}
		os.Exit(0)
	}

	// Get Together API key
	togetherKey, err := GetTogetherAPIKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Filter documents using Together AI
	fmt.Printf("Filtering documents with query: %s\n", query)
	fmt.Println("Sending to Together AI...")

	filter := NewDocumentFilter(togetherKey)
	matchingNames, err := filter.FilterDocuments(documents, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error filtering documents: %v\n", err)
		os.Exit(1)
	}

	// Find full document info for matches
	matchingDocuments := FindMatchingDocuments(documents, matchingNames)

	if len(matchingDocuments) == 0 {
		fmt.Println("\nNo documents matched your query.")
		os.Exit(0)
	}

	fmt.Printf("\nFound %d matching document(s):\n", len(matchingDocuments))
	for index, document := range matchingDocuments {
		fmt.Printf("  %d. %s\n", index+1, document.Name)
	}

	// If multiple matches, let user select or refine
	if len(matchingDocuments) > 1 {
		fmt.Println("\nMultiple documents match. Options:")
		fmt.Println("  - Enter a number to export that document")
		fmt.Println("  - Enter 'all' to export all matching documents")
		fmt.Println("  - Enter 'q' to quit")
		fmt.Print("\nYour choice: ")

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "q" {
			fmt.Println("Exiting.")
			os.Exit(0)
		}

		if input == "all" {
			// Export all matching documents
			for _, document := range matchingDocuments {
				exportDocument(driveClient, document, *outputDirectory, format)
			}
		} else {
			// Try to parse as number
			selection, err := strconv.Atoi(input)
			if err != nil || selection < 1 || selection > len(matchingDocuments) {
				fmt.Println("Invalid selection.")
				os.Exit(1)
			}
			exportDocument(driveClient, matchingDocuments[selection-1], *outputDirectory, format)
		}
	} else {
		// Single match - export directly
		exportDocument(driveClient, matchingDocuments[0], *outputDirectory, format)
	}
}

func exportDocument(client *DriveClient, document GoogleDocument, outputDirectory string, format ExportFormat) {
	filename := SanitizeFilename(document.Name) + format.Extension
	outputPath := filename
	if outputDirectory != "." && outputDirectory != "" {
		outputPath = outputDirectory + "/" + filename
	}

	fmt.Printf("\nExporting: %s -> %s\n", document.Name, outputPath)

	err := client.ExportDocument(document.ID, outputPath, format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error exporting %s: %v\n", document.Name, err)
		return
	}

	fmt.Println("Export complete!")
}
