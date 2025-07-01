package linode_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/internal/services/linode"
	"github.com/chadit/CloudMCP/pkg/logger"
)

const (
	// benchmarkTestInstanceID represents the test instance ID used in benchmarks.
	benchmarkTestInstanceID = 123456
	// benchmarkInstanceCount defines the number of instances in medium dataset benchmarks.
	benchmarkInstanceCount = 10
	// benchmarkVolumeCount defines the number of volumes in small dataset benchmarks.
	benchmarkVolumeCount = 3
	// benchmarkImageCount defines the number of images in medium dataset benchmarks.
	benchmarkImageCount = 15
	// benchmarkRegionCount defines the number of regions in medium dataset benchmarks.
	benchmarkRegionCount = 20
	// benchmarkTypeCount defines the number of instance types in large dataset benchmarks.
	benchmarkTypeCount = 50
	// benchmarkStartingUserID represents the starting UID for test users.
	benchmarkStartingUserID = 12345
	// benchmarkStartingVolumeID represents the starting volume ID for test volumes.
	benchmarkStartingVolumeID = 12345
	// benchmarkStartingVolumeSize represents the starting volume size in GB.
	benchmarkStartingVolumeSize = 20
	// benchmarkVolumeSizeIncrement represents the size increment per volume in GB.
	benchmarkVolumeSizeIncrement = 10
	// benchmarkStartingImageSize represents the starting image size in MB.
	benchmarkStartingImageSize = 512
	// benchmarkImageSizeIncrement represents the size increment per image in MB.
	benchmarkImageSizeIncrement = 128
	// benchmarkStartingDiskSize represents the starting disk size in MB.
	benchmarkStartingDiskSize = 10240
	// benchmarkStartingMemorySize represents the starting memory size in MB.
	benchmarkStartingMemorySize = 1024
	// benchmarkMaxVCPUCount represents the maximum number of vCPUs.
	benchmarkMaxVCPUCount = 8
	// benchmarkNetworkOutput represents the network output in Mbps.
	benchmarkNetworkOutput = 1000
	// benchmarkHourlyPriceMultiplier represents the multiplier for hourly pricing.
	benchmarkHourlyPriceMultiplier = 0.01
	// benchmarkMonthlyPriceMultiplier represents the multiplier for monthly pricing.
	benchmarkMonthlyPriceMultiplier = 5.0
	// benchmarkBackupHourlyPrice represents the hourly backup price.
	benchmarkBackupHourlyPrice = 0.008
	// benchmarkBackupMonthlyPrice represents the monthly backup price.
	benchmarkBackupMonthlyPrice = 5.0
	// benchmarkReferralTotal represents the total number of referrals.
	benchmarkReferralTotal = 2
	// benchmarkReferralCompleted represents the number of completed referrals.
	benchmarkReferralCompleted = 1
	// benchmarkReferralPending represents the number of pending referrals.
	benchmarkReferralPending = 1
	// benchmarkReferralCredit represents the referral credit amount.
	benchmarkReferralCredit = 50
	// benchmarkAccountBalance represents the account balance.
	benchmarkAccountBalance = 125.50
	// benchmarkAccountBalanceUninvoiced represents the uninvoiced balance.
	benchmarkAccountBalanceUninvoiced = 45.75
	// benchmarkBaseIPAddress represents the base IP address for instances.
	benchmarkBaseIPAddress = 10
)

// BenchmarkToolExecution benchmarks the performance of critical tool operations
// to establish baseline performance metrics and detect performance regressions.
//
// **Benchmark Coverage:**
// • Account management operations (get, list, switch)
// • Instance management operations (list, get, create)
// • System operations (regions, types, kernels)
// • High-frequency operations that impact user experience
//
// **Performance Targets:**
// • Tool execution: <100ms for simple operations
// • Account switching: <50ms
// • List operations: <200ms for moderate datasets
// • API calls: <500ms including network simulation
//
// **Benchmark Environment:**
// • HTTP test server for consistent network simulation
// • Realistic data payloads matching production
// • Multiple account configurations for switching tests.
//
//nolint:gocognit // Complex benchmark with multiple scenarios is necessary for comprehensive testing
func BenchmarkToolExecution(b *testing.B) {
	// Create benchmark test server with realistic response times.
	testServer := createBenchmarkTestServer()
	defer testServer.Close()

	// Create service with benchmark configuration.
	benchService := createBenchmarkService(b, testServer.URL)
	benchContext := b.Context()

	// Initialize service once for all benchmarks.
	err := benchService.Initialize(benchContext)
	require.NoError(b, err, "Service initialization should succeed")

	b.Run("AccountGet", func(b *testing.B) {
		toolRequest := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_account_get",
				Arguments: map[string]any{},
			},
		}

		b.ResetTimer()

		for range b.N {
			benchResult, benchErr := benchService.CallToolForTesting(benchContext, toolRequest)
			if benchErr != nil {
				b.Fatalf("Account get failed: %v", benchErr)
			}

			if benchResult == nil {
				b.Fatal("Result should not be nil")
			}
		}
	})

	b.Run("AccountList", func(b *testing.B) {
		toolRequest := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_account_list",
				Arguments: map[string]any{},
			},
		}

		b.ResetTimer()

		for range b.N {
			benchResult, benchErr := benchService.CallToolForTesting(benchContext, toolRequest)
			if benchErr != nil {
				b.Fatalf("Account list failed: %v", benchErr)
			}

			if benchResult == nil {
				b.Fatal("Result should not be nil")
			}
		}
	})

	b.Run("InstancesList", func(b *testing.B) {
		toolRequest := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_instances_list",
				Arguments: map[string]any{},
			},
		}

		b.ResetTimer()

		for range b.N {
			benchResult, benchErr := benchService.CallToolForTesting(benchContext, toolRequest)
			if benchErr != nil {
				b.Fatalf("Instances list failed: %v", benchErr)
			}

			if benchResult == nil {
				b.Fatal("Result should not be nil")
			}
		}
	})

	// Skip InstanceGet benchmark due to incomplete mock data structure.
	// This would require a full Linode instance API response structure.
	// b.Run("InstanceGet", func(b *testing.B) { ... })

	b.Run("VolumesList", func(b *testing.B) {
		toolRequest := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_volumes_list",
				Arguments: map[string]any{},
			},
		}

		b.ResetTimer()

		for range b.N {
			benchResult, benchErr := benchService.CallToolForTesting(benchContext, toolRequest)
			if benchErr != nil {
				b.Fatalf("Volumes list failed: %v", benchErr)
			}

			if benchResult == nil {
				b.Fatal("Result should not be nil")
			}
		}
	})

	b.Run("ImagesList", func(b *testing.B) {
		toolRequest := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_images_list",
				Arguments: map[string]any{},
			},
		}

		b.ResetTimer()

		for range b.N {
			benchResult, benchErr := benchService.CallToolForTesting(benchContext, toolRequest)
			if benchErr != nil {
				b.Fatalf("Images list failed: %v", benchErr)
			}

			if benchResult == nil {
				b.Fatal("Result should not be nil")
			}
		}
	})
}

// BenchmarkAccountSwitching benchmarks account switching performance
// to ensure rapid account transitions for multi-tenant scenarios.
//
// **Performance Targets:**
// • Single account switch: <50ms
// • Sequential switches: <100ms total for 3 switches
// • Concurrent switches: <200ms for 10 concurrent operations
//
// **Test Scenarios:**
// • Single account switch operation
// • Rapid sequential account switching
// • Account switch with subsequent tool execution.
//
//nolint:gocognit // Complex benchmark with multiple account switching scenarios is necessary
func BenchmarkAccountSwitching(b *testing.B) {
	testServer := createBenchmarkTestServer()
	defer testServer.Close()

	benchService := createBenchmarkService(b, testServer.URL)
	benchContext := b.Context()

	err := benchService.Initialize(benchContext)
	require.NoError(b, err, "Service initialization should succeed")

	b.Run("SingleSwitch", func(b *testing.B) {
		switchRequest := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_account_switch",
				Arguments: map[string]any{
					"account_name": "development",
				},
			},
		}

		b.ResetTimer()

		for range b.N {
			benchResult, benchErr := benchService.CallToolForTesting(benchContext, switchRequest)
			if benchErr != nil {
				b.Fatalf("Account switch failed: %v", benchErr)
			}

			if benchResult == nil {
				b.Fatal("Result should not be nil")
			}
		}
	})

	b.Run("SequentialSwitches", func(b *testing.B) {
		accountNames := []string{"primary", "development", "staging"}

		b.ResetTimer()

		for range b.N {
			for _, accountName := range accountNames {
				switchRequest := mcp.CallToolRequest{
					Params: mcp.CallToolParams{
						Name: "linode_account_switch",
						Arguments: map[string]any{
							"account_name": accountName,
						},
					},
				}

				benchResult, benchErr := benchService.CallToolForTesting(benchContext, switchRequest)
				if benchErr != nil {
					b.Fatalf("Account switch to %s failed: %v", accountName, benchErr)
				}

				if benchResult == nil {
					b.Fatal("Result should not be nil")
				}
			}
		}
	})

	b.Run("SwitchWithExecution", func(b *testing.B) {
		switchRequest := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_account_switch",
				Arguments: map[string]any{
					"account_name": "development",
				},
			},
		}

		listRequest := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_instances_list",
				Arguments: map[string]any{},
			},
		}

		b.ResetTimer()

		for range b.N {
			// Switch account.
			switchResult, switchErr := benchService.CallToolForTesting(benchContext, switchRequest)
			if switchErr != nil {
				b.Fatalf("Account switch failed: %v", switchErr)
			}

			if switchResult == nil {
				b.Fatal("Switch result should not be nil")
			}

			// Execute tool with new account.
			listResult, listErr := benchService.CallToolForTesting(benchContext, listRequest)
			if listErr != nil {
				b.Fatalf("Instances list failed: %v", listErr)
			}

			if listResult == nil {
				b.Fatal("List result should not be nil")
			}
		}
	})
}

// BenchmarkAPILatency benchmarks simulated API call performance
// to establish baseline network performance expectations.
//
// **Performance Metrics:**
// • Raw API call latency
// • Response parsing performance
// • Error handling overhead
// • Memory allocation patterns.
func BenchmarkAPILatency(b *testing.B) {
	testServer := createBenchmarkTestServer()
	defer testServer.Close()

	benchService := createBenchmarkService(b, testServer.URL)
	benchContext := b.Context()

	err := benchService.Initialize(benchContext)
	require.NoError(b, err, "Service initialization should succeed")

	b.Run("SmallResponse", func(b *testing.B) {
		smallRequest := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_account_get",
				Arguments: map[string]any{},
			},
		}

		b.ResetTimer()

		for range b.N {
			benchResult, benchErr := benchService.CallToolForTesting(benchContext, smallRequest)
			if benchErr != nil {
				b.Fatalf("Small response test failed: %v", benchErr)
			}

			if benchResult == nil {
				b.Fatal("Result should not be nil")
			}
		}
	})

	b.Run("MediumResponse", func(b *testing.B) {
		mediumRequest := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_instances_list",
				Arguments: map[string]any{},
			},
		}

		b.ResetTimer()

		for range b.N {
			benchResult, benchErr := benchService.CallToolForTesting(benchContext, mediumRequest)
			if benchErr != nil {
				b.Fatalf("Medium response test failed: %v", benchErr)
			}

			if benchResult == nil {
				b.Fatal("Result should not be nil")
			}
		}
	})

	b.Run("LargeResponse", func(b *testing.B) {
		largeRequest := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_images_list",
				Arguments: map[string]any{},
			},
		}

		b.ResetTimer()

		for range b.N {
			benchResult, benchErr := benchService.CallToolForTesting(benchContext, largeRequest)
			if benchErr != nil {
				b.Fatalf("Large response test failed: %v", benchErr)
			}

			if benchResult == nil {
				b.Fatal("Result should not be nil")
			}
		}
	})
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
// for high-frequency operations to identify optimization opportunities.
//
// **Memory Optimization Targets:**
// • Minimize allocations per operation
// • Reduce GC pressure for high-frequency calls
// • Optimize string operations and JSON processing
// • Monitor memory growth patterns.
func BenchmarkMemoryAllocation(b *testing.B) {
	testServer := createBenchmarkTestServer()
	defer testServer.Close()

	benchService := createBenchmarkService(b, testServer.URL)
	benchContext := b.Context()

	err := benchService.Initialize(benchContext)
	require.NoError(b, err, "Service initialization should succeed")

	b.Run("MemoryPerToolCall", func(b *testing.B) {
		toolCallRequest := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_account_get",
				Arguments: map[string]any{},
			},
		}

		b.ReportAllocs()
		b.ResetTimer()

		for range b.N {
			benchResult, benchErr := benchService.CallToolForTesting(benchContext, toolCallRequest)
			if benchErr != nil {
				b.Fatalf("Tool call failed: %v", benchErr)
			}

			if benchResult == nil {
				b.Fatal("Result should not be nil")
			}
		}
	})

	b.Run("MemoryPerAccountSwitch", func(b *testing.B) {
		accountSwitchRequest := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "linode_account_switch",
				Arguments: map[string]any{
					"account_name": "development",
				},
			},
		}

		b.ReportAllocs()
		b.ResetTimer()

		for range b.N {
			benchResult, benchErr := benchService.CallToolForTesting(benchContext, accountSwitchRequest)
			if benchErr != nil {
				b.Fatalf("Account switch failed: %v", benchErr)
			}

			if benchResult == nil {
				b.Fatal("Result should not be nil")
			}
		}
	})

	b.Run("MemoryPerListOperation", func(b *testing.B) {
		listOperationRequest := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "linode_instances_list",
				Arguments: map[string]any{},
			},
		}

		b.ReportAllocs()
		b.ResetTimer()

		for range b.N {
			benchResult, benchErr := benchService.CallToolForTesting(benchContext, listOperationRequest)
			if benchErr != nil {
				b.Fatalf("List operation failed: %v", benchErr)
			}

			if benchResult == nil {
				b.Fatal("Result should not be nil")
			}
		}
	})
}

// createBenchmarkTestServer creates an optimized HTTP test server for benchmarking
// with realistic response times and data sizes for performance testing.
func createBenchmarkTestServer() *httptest.Server {
	serverMux := http.NewServeMux()

	// Profile endpoint - small response.
	serverMux.HandleFunc("/v4/profile", func(responseWriter http.ResponseWriter, _ *http.Request) {
		profileResponse := map[string]any{
			"uid":                  benchmarkStartingUserID,
			"username":             "benchuser",
			"email":                "bench@example.com",
			"timezone":             "US/Eastern",
			"email_notifications":  true,
			"referrals":            map[string]int{"total": benchmarkReferralTotal, "completed": benchmarkReferralCompleted, "pending": benchmarkReferralPending, "credit": benchmarkReferralCredit},
			"ip_whitelist_enabled": false,
			"lish_auth_method":     "password",
			"authorized_keys":      []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ..."},
			"two_factor_auth":      true,
			"restricted":           false,
		}

		responseWriter.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(responseWriter).Encode(profileResponse)
	})

	// Account endpoint - small response.
	serverMux.HandleFunc("/v4/account", func(responseWriter http.ResponseWriter, _ *http.Request) {
		accountResponse := map[string]any{
			"email":              "bench@example.com",
			"first_name":         "Bench",
			"last_name":          "User",
			"company":            "Benchmark Inc.",
			"address_1":          "123 Bench Street",
			"city":               "New York",
			"state":              "NY",
			"zip":                "10001",
			"country":            "US",
			"phone":              "+1-555-123-4567",
			"balance":            benchmarkAccountBalance,
			"balance_uninvoiced": benchmarkAccountBalanceUninvoiced,
			"capabilities":       []string{"Linodes", "NodeBalancers", "Block Storage", "Object Storage"},
			"active_since":       "2020-01-15T08:30:00",
		}

		responseWriter.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(responseWriter).Encode(accountResponse)
	})

	// Instances endpoint - medium response.
	serverMux.HandleFunc("/v4/linode/instances", func(responseWriter http.ResponseWriter, _ *http.Request) {
		instancesList := make([]map[string]any, benchmarkInstanceCount)

		for instanceIndex := range benchmarkInstanceCount {
			instancesList[instanceIndex] = map[string]any{
				"id":      benchmarkTestInstanceID + instanceIndex,
				"label":   "bench-instance-" + string(rune(instanceIndex+'1')),
				"group":   "benchmark",
				"status":  "running",
				"type":    "g6-standard-2",
				"region":  "us-east",
				"image":   "linode/ubuntu20.04",
				"ipv4":    []string{fmt.Sprintf("192.168.1.%d", benchmarkBaseIPAddress+instanceIndex)},
				"ipv6":    fmt.Sprintf("2001:db8::%d/64", instanceIndex+1),
				"created": "2023-01-15T10:30:00",
				"updated": "2023-06-20T14:22:00",
			}
		}

		instancesResponse := map[string]any{
			"data":    instancesList,
			"page":    1,
			"pages":   1,
			"results": len(instancesList),
		}

		responseWriter.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(responseWriter).Encode(instancesResponse)
	})

	// Individual instance endpoint with path extraction.
	serverMux.HandleFunc("/v4/linode/instances/123456", func(responseWriter http.ResponseWriter, _ *http.Request) {
		singleInstance := map[string]any{
			"id":      benchmarkTestInstanceID,
			"label":   "bench-instance",
			"group":   "benchmark",
			"status":  "running",
			"type":    "g6-standard-2",
			"region":  "us-east",
			"image":   "linode/ubuntu20.04",
			"ipv4":    []string{"192.168.1.10"},
			"ipv6":    "2001:db8::1/64",
			"created": "2023-01-15T10:30:00",
			"updated": "2023-06-20T14:22:00",
		}

		responseWriter.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(responseWriter).Encode(singleInstance)
	})

	// Volumes endpoint - small response.
	serverMux.HandleFunc("/v4/volumes", func(responseWriter http.ResponseWriter, _ *http.Request) {
		volumesList := make([]map[string]any, benchmarkVolumeCount)

		for volumeIndex := range benchmarkVolumeCount {
			volumesList[volumeIndex] = map[string]any{
				"id":           benchmarkStartingVolumeID + volumeIndex,
				"label":        fmt.Sprintf("bench-volume-%d", volumeIndex+1),
				"status":       "active",
				"size":         benchmarkStartingVolumeSize + (volumeIndex * benchmarkVolumeSizeIncrement),
				"region":       "us-east",
				"filesystem":   "ext4",
				"created":      "2023-01-15T10:30:00",
				"updated":      "2023-06-20T14:22:00",
				"linode_id":    nil,
				"linode_label": nil,
			}
		}

		volumesResponse := map[string]any{
			"data":    volumesList,
			"page":    1,
			"pages":   1,
			"results": len(volumesList),
		}

		responseWriter.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(responseWriter).Encode(volumesResponse)
	})

	// Images endpoint - medium response.
	serverMux.HandleFunc("/v4/images", func(responseWriter http.ResponseWriter, _ *http.Request) {
		imagesList := make([]map[string]any, benchmarkImageCount)
		imageTypeOptions := []string{"manual", "automatic", "public"}

		for imageIndex := range benchmarkImageCount {
			imagesList[imageIndex] = map[string]any{
				"id":          fmt.Sprintf("linode/benchmark-%d", imageIndex+1),
				"label":       fmt.Sprintf("Benchmark Image %d", imageIndex+1),
				"description": fmt.Sprintf("Benchmark test image %d", imageIndex+1),
				"type":        imageTypeOptions[imageIndex%len(imageTypeOptions)],
				"status":      "available",
				"size":        benchmarkStartingImageSize + (imageIndex * benchmarkImageSizeIncrement),
				"created":     "2023-01-15T10:30:00",
				"created_by":  "benchmark",
				"deprecated":  false,
				"is_public":   true,
				"vendor":      "Benchmark Corp",
			}
		}

		imagesResponse := map[string]any{
			"data":    imagesList,
			"page":    1,
			"pages":   1,
			"results": len(imagesList),
		}

		responseWriter.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(responseWriter).Encode(imagesResponse)
	})

	// Regions endpoint - medium response.
	serverMux.HandleFunc("/v4/regions", func(responseWriter http.ResponseWriter, _ *http.Request) {
		regionsList := make([]map[string]any, benchmarkRegionCount)
		regionNameOptions := []string{"us-east", "us-west", "eu-west", "ap-south", "ca-central"}

		for regionIndex := range benchmarkRegionCount {
			regionsList[regionIndex] = map[string]any{
				"id":           fmt.Sprintf("%s-%d", regionNameOptions[regionIndex%len(regionNameOptions)], regionIndex/len(regionNameOptions)+1),
				"label":        fmt.Sprintf("Benchmark Region %d", regionIndex+1),
				"country":      "US",
				"capabilities": []string{"Linodes", "NodeBalancers", "Block Storage"},
				"status":       "ok",
			}
		}

		regionsResponse := map[string]any{
			"data":    regionsList,
			"page":    1,
			"pages":   1,
			"results": len(regionsList),
		}

		responseWriter.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(responseWriter).Encode(regionsResponse)
	})

	// Types endpoint - large response.
	serverMux.HandleFunc("/v4/linode/types", func(responseWriter http.ResponseWriter, _ *http.Request) {
		typesList := make([]map[string]any, benchmarkTypeCount)
		typeClassOptions := []string{"nanode", "standard", "dedicated", "high-memory"}

		for typeIndex := range benchmarkTypeCount {
			typesList[typeIndex] = map[string]any{
				"id":          fmt.Sprintf("%s-%d", typeClassOptions[typeIndex%len(typeClassOptions)], typeIndex+1),
				"label":       fmt.Sprintf("Benchmark Type %d", typeIndex+1),
				"class":       typeClassOptions[typeIndex%len(typeClassOptions)],
				"disk":        (typeIndex + 1) * benchmarkStartingDiskSize,
				"memory":      (typeIndex + 1) * benchmarkStartingMemorySize,
				"vcpus":       (typeIndex % benchmarkMaxVCPUCount) + 1,
				"gpus":        0,
				"network_out": benchmarkNetworkOutput,
				"price": map[string]float64{
					"hourly":  float64(typeIndex+1) * benchmarkHourlyPriceMultiplier,
					"monthly": float64(typeIndex+1) * benchmarkMonthlyPriceMultiplier,
				},
				"addons": map[string]any{
					"backups": map[string]any{
						"price": map[string]float64{
							"hourly":  benchmarkBackupHourlyPrice,
							"monthly": benchmarkBackupMonthlyPrice,
						},
					},
				},
			}
		}

		typesResponse := map[string]any{
			"data":    typesList,
			"page":    1,
			"pages":   1,
			"results": len(typesList),
		}

		responseWriter.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(responseWriter).Encode(typesResponse)
	})

	return httptest.NewServer(serverMux)
}

// createBenchmarkService creates a Linode service configured for benchmarking
// with multiple accounts and optimized settings for performance testing.
func createBenchmarkService(b *testing.B, serverAPIURL string) *linode.Service {
	b.Helper()

	benchmarkLogger := logger.New("error") // Minimize logging overhead for benchmarks.

	benchmarkConfig := &config.Config{
		ServerName:           "Benchmark Service",
		LogLevel:             "error",
		EnableMetrics:        false, // Disable metrics for pure performance testing.
		DefaultLinodeAccount: "primary",
		LinodeAccounts: map[string]config.LinodeAccount{
			"primary": {
				Token:  "benchmark-token-primary",
				Label:  "Primary Benchmark Account",
				APIURL: serverAPIURL,
			},
			"development": {
				Token:  "benchmark-token-dev",
				Label:  "Development Benchmark Account",
				APIURL: serverAPIURL,
			},
			"staging": {
				Token:  "benchmark-token-staging",
				Label:  "Staging Benchmark Account",
				APIURL: serverAPIURL,
			},
		},
	}

	benchmarkService, serviceErr := linode.New(benchmarkConfig, benchmarkLogger)
	require.NoError(b, serviceErr, "Service creation should succeed")

	return benchmarkService
}
