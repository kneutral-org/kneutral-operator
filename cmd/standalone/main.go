package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/kneutral-org/kneutral-operator/internal/api"
	"github.com/kneutral-org/kneutral-operator/internal/mock"
)

func main() {
	var apiAddr string
	var mockData bool

	flag.StringVar(&apiAddr, "api-bind-address", ":8090", "The address the API server binds to.")
	flag.BoolVar(&mockData, "mock-data", true, "Enable mock data for testing without Kubernetes")
	flag.Parse()

	fmt.Printf("🚀 Starting Kneutral Operator API in standalone mode\n")
	fmt.Printf("📍 API Address: %s\n", apiAddr)
	fmt.Printf("🎭 Mock Data: %v\n", mockData)

	// Create mock client for standalone testing
	client := mock.NewMockClient()

	if mockData {
		// Pre-populate with example data
		mock.PopulateExampleData(client)
		fmt.Printf("✅ Loaded mock data with example AlertRules\n")
	}

	// Start API server
	apiServer := api.NewServer(client, apiAddr)
	fmt.Printf("🌐 API Documentation: http://localhost%s/docs\n", apiAddr)
	fmt.Printf("📊 Health Check: http://localhost%s/health\n", apiAddr)
	fmt.Printf("🔍 List AlertRules: http://localhost%s/api/v1/alertrules\n", apiAddr)

	if err := apiServer.Start(); err != nil {
		log.Fatalf("Failed to start API server: %v", err)
	}
}