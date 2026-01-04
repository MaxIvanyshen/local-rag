package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/MaxIvanyshen/local-rag/config"
	"github.com/MaxIvanyshen/local-rag/db"
	"github.com/MaxIvanyshen/local-rag/service"
)

func main() {
	cfg := config.GetConfig(context.Background())

	url := fmt.Sprintf("%s:%d", cfg.Extensions.Host, cfg.Port)

	var serverURL string
	flag.StringVar(&serverURL, "url", url, "URL of the local RAG service")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: rag <command> [args...]")
		fmt.Println("Commands:")
		fmt.Println("  search <query>           - Search for documents")
		fmt.Println("  process <filename>       - Process a single document")
		fmt.Println("  delete <name>            - Delete a document by name")
		fmt.Println("  batch <filename>...      - Process multiple documents")
		os.Exit(1)
	}

	command := args[0]
	switch command {
	case "search":
		if len(args) < 2 {
			fmt.Println("Usage: rag search <query>")
			os.Exit(1)
		}
		query := args[1]
		search(serverURL, query)
	case "process":
		if len(args) < 2 {
			fmt.Println("Usage: rag process <filename>")
			os.Exit(1)
		}
		filename := args[1]
		process(serverURL, filename)
	case "delete":
		if len(args) < 2 {
			fmt.Println("Usage: rag delete <name>")
			os.Exit(1)
		}
		name := args[1]
		deleteDoc(serverURL, name)
	case "batch":
		if len(args) < 2 {
			fmt.Println("Usage: rag batch <filename>...")
			os.Exit(1)
		}
		filenames := args[1:]
		batchProcess(serverURL, filenames)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func search(serverURL, query string) {
	req := service.SearchRequest{Query: query}
	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/api/search", "application/json", bytes.NewBuffer(body))
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "dial tcp") {
			fmt.Printf("Error: Service appears to be not running. Please start the server first.\n")
		} else {
			fmt.Printf("Error making request: %v\n", err)
		}
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Server error: %s\n", resp.Status)
		os.Exit(1)
	}

	var results []db.SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Println("No results found.")
		return
	}

	fmt.Println("Search Results:")
	for _, result := range results {
		fmt.Printf("- Document: %s\n", result.DocumentName)
		fmt.Printf("  Content: %s\n", result.Content)
		fmt.Printf("  Distance: %.4f\n", result.Distance)
		fmt.Println()
	}
}

func process(serverURL, filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	req := service.ProcessDocumentRequest{
		DocumentName: filename,
		DocumentData: data,
	}
	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/api/process_document", "application/json", bytes.NewBuffer(body))
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "dial tcp") {
			fmt.Printf("Error: Service appears to be not running. Please start the server first.\n")
		} else {
			fmt.Printf("Error making request: %v\n", err)
		}
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Server error: %s - %s\n", resp.Status, string(body))
		os.Exit(1)
	}

	var success service.SuccessResponse
	if err := json.NewDecoder(resp.Body).Decode(&success); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		os.Exit(1)
	}

	if success.Success {
		fmt.Println("Document processed successfully.")
	} else {
		fmt.Println("Document processing failed.")
		os.Exit(1)
	}
}

func deleteDoc(serverURL, name string) {
	req := service.DeleteDocumentRequest{
		DocumentName: name,
	}
	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/api/delete_document", "application/json", bytes.NewBuffer(body))
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "dial tcp") {
			fmt.Printf("Error: Service appears to be not running. Please start the server first.\n")
		} else {
			fmt.Printf("Error making request: %v\n", err)
		}
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Server error: %s - %s\n", resp.Status, string(body))
		os.Exit(1)
	}

	var success service.SuccessResponse
	if err := json.NewDecoder(resp.Body).Decode(&success); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		os.Exit(1)
	}

	if success.Success {
		fmt.Println("Document deleted successfully.")
	} else {
		fmt.Println("Document deletion failed.")
		os.Exit(1)
	}
}

func batchProcess(serverURL string, filenames []string) {
	var reqs []*service.ProcessDocumentRequest
	for _, filename := range filenames {
		data, err := os.ReadFile(filename)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", filename, err)
			os.Exit(1)
		}
		reqs = append(reqs, &service.ProcessDocumentRequest{
			DocumentName: filename,
			DocumentData: data,
		})
	}

	body, err := json.Marshal(service.BatchProcessDocumentsRequest{Documents: reqs})
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/api/batch_process_documents", "application/json", bytes.NewBuffer(body))
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "dial tcp") {
			fmt.Printf("Error: Service appears to be not running. Please start the server first.\n")
		} else {
			fmt.Printf("Error making request: %v\n", err)
		}
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Server error: %s - %s\n", resp.Status, string(body))
		os.Exit(1)
	}

	var success service.SuccessResponse
	if err := json.NewDecoder(resp.Body).Decode(&success); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		os.Exit(1)
	}

	if success.Success {
		fmt.Printf("Batch processed %d documents successfully.\n", len(filenames))
	} else {
		fmt.Println("Batch processing failed.")
		os.Exit(1)
	}
}
