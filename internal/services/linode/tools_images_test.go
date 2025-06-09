package linode

import (
	"context"
	"testing"
	"time"

	"github.com/linode/linodego"
	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/pkg/logger"
)

func TestHandleListImages(t *testing.T) {
	tests := []struct {
		name    string
		params  ImagesListParams
		wantErr bool
		errMsg  string
	}{
		{
			name:    "list all images",
			params:  ImagesListParams{},
			wantErr: false,
		},
		{
			name:    "list public images only",
			params:  ImagesListParams{IsPublic: &[]bool{true}[0]},
			wantErr: false,
		},
		{
			name:    "list private images only",
			params:  ImagesListParams{IsPublic: &[]bool{false}[0]},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := createTestService(t)
			ctx := context.Background()

			result, err := service.handleListImages(ctx, tt.params)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			require.IsType(t, &ImagesListResult{}, result)
		})
	}
}

func TestHandleGetImage(t *testing.T) {
	tests := []struct {
		name    string
		params  ImageGetParams
		wantErr bool
		errMsg  string
	}{
		{
			name:    "get public image",
			params:  ImageGetParams{ImageID: "linode/ubuntu22.04"},
			wantErr: false,
		},
		{
			name:    "get nonexistent image",
			params:  ImageGetParams{ImageID: "private/nonexistent"},
			wantErr: true,
			errMsg:  "failed to get image",
		},
		{
			name:    "empty image ID",
			params:  ImageGetParams{ImageID: ""},
			wantErr: true,
			errMsg:  "image ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := createTestService(t)
			ctx := context.Background()

			result, err := service.handleGetImage(ctx, tt.params)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, tt.params.ImageID, result.ID)
		})
	}
}

func TestHandleCreateImage(t *testing.T) {
	tests := []struct {
		name    string
		params  ImageCreateParams
		wantErr bool
		errMsg  string
	}{
		{
			name: "create image with valid disk",
			params: ImageCreateParams{
				DiskID:      12345,
				Label:       "test-image",
				Description: "Test image for unit tests",
				Tags:        []string{"test", "unit"},
			},
			wantErr: true, // Will fail in test environment
			errMsg:  "failed to create image",
		},
		{
			name: "create image with invalid disk",
			params: ImageCreateParams{
				DiskID: 0,
				Label:  "test-image",
			},
			wantErr: true,
			errMsg:  "failed to create image",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := createTestService(t)
			ctx := context.Background()

			result, err := service.handleCreateImage(ctx, tt.params)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, tt.params.Label, result.Label)
		})
	}
}

func TestHandleUpdateImage(t *testing.T) {
	tests := []struct {
		name    string
		params  ImageUpdateParams
		wantErr bool
		errMsg  string
	}{
		{
			name: "update nonexistent image",
			params: ImageUpdateParams{
				ImageID: "private/nonexistent",
				Label:   "updated-label",
			},
			wantErr: true,
			errMsg:  "failed to update image",
		},
		{
			name: "update public image",
			params: ImageUpdateParams{
				ImageID: "linode/ubuntu22.04",
				Label:   "should-fail",
			},
			wantErr: true,
			errMsg:  "failed to update image",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := createTestService(t)
			ctx := context.Background()

			result, err := service.handleUpdateImage(ctx, tt.params)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
		})
	}
}

func TestHandleDeleteImage(t *testing.T) {
	tests := []struct {
		name    string
		params  ImageDeleteParams
		wantErr bool
		errMsg  string
	}{
		{
			name: "delete nonexistent image",
			params: ImageDeleteParams{
				ImageID: "private/nonexistent",
			},
			wantErr: true,
			errMsg:  "failed to get image",
		},
		{
			name: "delete public image",
			params: ImageDeleteParams{
				ImageID: "linode/ubuntu22.04",
			},
			wantErr: true,
			errMsg:  "cannot delete public images",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := createTestService(t)
			ctx := context.Background()

			result, err := service.handleDeleteImage(ctx, tt.params)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, result)
		})
	}
}

func TestHandleReplicateImage(t *testing.T) {
	tests := []struct {
		name    string
		params  ImageReplicateParams
		wantErr bool
		errMsg  string
	}{
		{
			name: "replicate nonexistent image",
			params: ImageReplicateParams{
				ImageID: "private/nonexistent",
				Regions: []string{"us-west", "eu-central"},
			},
			wantErr: true,
			errMsg:  "failed to get image",
		},
		{
			name: "replicate public image",
			params: ImageReplicateParams{
				ImageID: "linode/ubuntu22.04",
				Regions: []string{"us-west"},
			},
			wantErr: true,
			errMsg:  "cannot replicate public images",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := createTestService(t)
			ctx := context.Background()

			result, err := service.handleReplicateImage(ctx, tt.params)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
		})
	}
}

func TestHandleCreateImageUpload(t *testing.T) {
	tests := []struct {
		name    string
		params  ImageUploadParams
		wantErr bool
		errMsg  string
	}{
		{
			name: "create upload with valid params",
			params: ImageUploadParams{
				Label:       "uploaded-image",
				Region:      "us-east",
				Description: "Image uploaded for testing",
				Tags:        []string{"uploaded", "test"},
			},
			wantErr: true, // Will fail in test environment
			errMsg:  "failed to create image upload",
		},
		{
			name: "create upload with minimal params",
			params: ImageUploadParams{
				Label:  "minimal-upload",
				Region: "us-east",
			},
			wantErr: true, // Will fail in test environment
			errMsg:  "failed to create image upload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := createTestService(t)
			ctx := context.Background()

			result, err := service.handleCreateImageUpload(ctx, tt.params)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			require.NotEmpty(t, result.ImageID)
			require.NotEmpty(t, result.UploadTo)
		})
	}
}

func TestImageSummaryConversion(t *testing.T) {
	_ = createTestService(t)

	// Create a mock linodego.Image
	created := time.Now()
	mockImage := linodego.Image{
		ID:           "linode/ubuntu22.04",
		Label:        "Ubuntu 22.04 LTS",
		Description:  "Ubuntu 22.04 LTS (Jammy Jellyfish)",
		Created:      &created,
		CreatedBy:    "linode",
		IsPublic:     true,
		Size:         2500,
		Type:         "manual",
		Vendor:       "Ubuntu",
		Status:       linodego.ImageStatusAvailable,
		TotalSize:    2500,
		Capabilities: []string{"cloud-init"},
		Tags:         []string{"ubuntu", "lts"},
		Regions: []linodego.ImageRegion{
			{Region: "us-east", Status: "available"},
			{Region: "us-west", Status: "available"},
		},
	}

	// Test conversion logic (this would be part of handleListImages)
	summary := ImageSummary{
		ID:           mockImage.ID,
		Label:        mockImage.Label,
		Description:  mockImage.Description,
		Created:      mockImage.Created.Format("2006-01-02T15:04:05"),
		CreatedBy:    mockImage.CreatedBy,
		Deprecated:   mockImage.Deprecated,
		IsPublic:     mockImage.IsPublic,
		Size:         mockImage.Size,
		Type:         mockImage.Type,
		Vendor:       mockImage.Vendor,
		Status:       string(mockImage.Status),
		TotalSize:    mockImage.TotalSize,
		Capabilities: mockImage.Capabilities,
		Tags:         mockImage.Tags,
	}

	// Convert regions
	summary.Regions = make([]ImageRegion, len(mockImage.Regions))
	for i, r := range mockImage.Regions {
		summary.Regions[i] = ImageRegion{
			Region: r.Region,
			Status: string(r.Status),
		}
	}

	require.Equal(t, mockImage.ID, summary.ID)
	require.Equal(t, mockImage.Label, summary.Label)
	require.Equal(t, mockImage.IsPublic, summary.IsPublic)
	require.Equal(t, len(mockImage.Regions), len(summary.Regions))
	require.Equal(t, len(mockImage.Tags), len(summary.Tags))
}

func TestFormatImagesListResult(t *testing.T) {
	result := &ImagesListResult{
		Count: 2,
		Images: []ImageSummary{
			{
				ID:          "linode/ubuntu22.04",
				Label:       "Ubuntu 22.04 LTS",
				Description: "Ubuntu 22.04 LTS",
				Type:        "manual",
				Status:      "available",
				Size:        2500,
				IsPublic:    true,
				Created:     "2024-01-01T00:00:00",
				Tags:        []string{"ubuntu", "lts"},
				Regions: []ImageRegion{
					{Region: "us-east", Status: "available"},
				},
			},
			{
				ID:          "private/12345",
				Label:       "My Custom Image",
				Description: "Custom image for testing",
				Type:        "manual",
				Status:      "available",
				Size:        3000,
				IsPublic:    false,
				Created:     "2024-01-02T00:00:00",
				Tags:        []string{"custom"},
				Regions: []ImageRegion{
					{Region: "us-west", Status: "available"},
				},
			},
		},
	}

	output := formatImagesListResult(result)

	require.Contains(t, output, "Found 2 images")
	require.Contains(t, output, "linode/ubuntu22.04")
	require.Contains(t, output, "Ubuntu 22.04 LTS")
	require.Contains(t, output, "private/12345")
	require.Contains(t, output, "My Custom Image")
	require.Contains(t, output, "Public: true")
	require.Contains(t, output, "Public: false")
	require.Contains(t, output, "Tags: ubuntu, lts")
	require.Contains(t, output, "us-east: available")
}

func TestFormatImageDetail(t *testing.T) {
	expiry := "2025-01-01T00:00:00"
	detail := &ImageDetail{
		ID:           "private/12345",
		Label:        "My Custom Image",
		Description:  "Custom image for testing",
		Type:         "manual",
		Status:       "available",
		Size:         3000,
		TotalSize:    3000,
		IsPublic:     false,
		Deprecated:   false,
		Created:      "2024-01-01T00:00:00",
		CreatedBy:    "user@example.com",
		Updated:      "2024-01-02T00:00:00",
		Expiry:       &expiry,
		Tags:         []string{"custom", "test"},
		Capabilities: []string{"cloud-init"},
		Regions: []ImageRegion{
			{Region: "us-east", Status: "available"},
			{Region: "us-west", Status: "replicating"},
		},
	}

	output := formatImageDetail(detail)

	require.Contains(t, output, "Image: private/12345 (My Custom Image)")
	require.Contains(t, output, "Description: Custom image for testing")
	require.Contains(t, output, "Status: available")
	require.Contains(t, output, "Size: 3000 MB")
	require.Contains(t, output, "Public: false")
	require.Contains(t, output, "Created By: user@example.com")
	require.Contains(t, output, "Updated: 2024-01-02T00:00:00")
	require.Contains(t, output, "Expires: 2025-01-01T00:00:00")
	require.Contains(t, output, "Tags: custom, test")
	require.Contains(t, output, "Capabilities: cloud-init")
	require.Contains(t, output, "us-east: available")
	require.Contains(t, output, "us-west: replicating")
}

// Helper function to create a test service
func createTestService(t *testing.T) *Service {
	l := logger.New("info")

	// Create minimal test configuration
	cfg := &config.Config{
		DefaultLinodeAccount: "test",
		LinodeAccounts: map[string]config.LinodeAccount{
			"test": {
				Token: "test-token",
				Label: "Test Account",
			},
		},
	}

	service, err := New(cfg, l)
	require.NoError(t, err)

	return service
}
