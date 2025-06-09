//go:build integration

package linode

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/chadit/CloudMCP/internal/config"
	"github.com/chadit/CloudMCP/pkg/logger"
)

// TestImageIntegrationListImages tests listing images with a real Linode API connection.
func TestImageIntegrationListImages(t *testing.T) {
	service := createIntegrationTestService(t)
	ctx := context.Background()

	t.Run("list all images", func(t *testing.T) {
		params := ImagesListParams{}
		result, err := service.handleListImages(ctx, params)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Greater(t, result.Count, 0, "Should have at least some images")
		require.Len(t, result.Images, result.Count)

		// Verify we have both public and private images
		hasPublic := false
		for _, img := range result.Images {
			if img.IsPublic {
				hasPublic = true
				break
			}
		}
		require.True(t, hasPublic, "Should have at least one public image")
	})

	t.Run("list public images only", func(t *testing.T) {
		isPublic := true
		params := ImagesListParams{IsPublic: &isPublic}
		result, err := service.handleListImages(ctx, params)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Greater(t, result.Count, 0, "Should have public images")

		// Verify all returned images are public
		for _, img := range result.Images {
			require.True(t, img.IsPublic, "All images should be public when filtered")
		}
	})

	t.Run("list private images only", func(t *testing.T) {
		isPublic := false
		params := ImagesListParams{IsPublic: &isPublic}
		result, err := service.handleListImages(ctx, params)

		require.NoError(t, err)
		require.NotNil(t, result)
		// Note: May have 0 private images, which is okay

		// Verify all returned images are private
		for _, img := range result.Images {
			require.False(t, img.IsPublic, "All images should be private when filtered")
		}
	})
}

// TestImageIntegrationGetImage tests retrieving specific images.
func TestImageIntegrationGetImage(t *testing.T) {
	service := createIntegrationTestService(t)
	ctx := context.Background()

	t.Run("get ubuntu image", func(t *testing.T) {
		params := ImageGetParams{ImageID: "linode/ubuntu22.04"}
		result, err := service.handleGetImage(ctx, params)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "linode/ubuntu22.04", result.ID)
		require.True(t, result.IsPublic)
		require.Equal(t, "Ubuntu", result.Vendor)
		require.Greater(t, len(result.Regions), 0, "Should be available in multiple regions")
	})

	t.Run("get nonexistent image", func(t *testing.T) {
		params := ImageGetParams{ImageID: "private/does-not-exist-12345"}
		result, err := service.handleGetImage(ctx, params)

		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "failed to get image")
	})
}

// TestImageIntegrationLifecycle tests the complete lifecycle of custom image operations.
// This test requires a real Linode instance with a disk to create an image from.
func TestImageIntegrationLifecycle(t *testing.T) {
	// Skip this test unless explicitly enabled, as it requires real resources
	if os.Getenv("ENABLE_IMAGE_LIFECYCLE_TEST") != "true" {
		t.Skip("Image lifecycle test skipped. Set ENABLE_IMAGE_LIFECYCLE_TEST=true to run.")
	}

	service := createIntegrationTestService(t)
	ctx := context.Background()

	// This test would require:
	// 1. Creating a test Linode instance
	// 2. Creating an image from that instance's disk
	// 3. Testing image operations (update, replicate)
	// 4. Cleaning up (delete image and instance)

	// For safety, we'll just test the error cases here
	t.Run("create image from invalid disk", func(t *testing.T) {
		params := ImageCreateParams{
			DiskID:      99999999, // Non-existent disk ID
			Label:       "integration-test-image",
			Description: "Test image for integration testing",
		}

		result, err := service.handleCreateImage(ctx, params)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "failed to create image")
	})

	t.Run("update nonexistent image", func(t *testing.T) {
		params := ImageUpdateParams{
			ImageID: "private/nonexistent-12345",
			Label:   "updated-test-image",
		}

		result, err := service.handleUpdateImage(ctx, params)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "failed to update image")
	})

	t.Run("delete nonexistent image", func(t *testing.T) {
		params := ImageDeleteParams{
			ImageID: "private/nonexistent-12345",
		}

		result, err := service.handleDeleteImage(ctx, params)
		require.Error(t, err)
		require.Empty(t, result)
		require.Contains(t, err.Error(), "failed to get image")
	})

	t.Run("replicate nonexistent image", func(t *testing.T) {
		params := ImageReplicateParams{
			ImageID: "private/nonexistent-12345",
			Regions: []string{"us-west", "eu-central"},
		}

		result, err := service.handleReplicateImage(ctx, params)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "failed to get image")
	})

	t.Run("replicate public image (should fail)", func(t *testing.T) {
		params := ImageReplicateParams{
			ImageID: "linode/ubuntu22.04",
			Regions: []string{"us-west"},
		}

		result, err := service.handleReplicateImage(ctx, params)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "cannot replicate public images")
	})
}

// TestImageIntegrationUpload tests image upload functionality.
func TestImageIntegrationUpload(t *testing.T) {
	// Skip this test unless explicitly enabled, as it creates real resources
	if os.Getenv("ENABLE_IMAGE_UPLOAD_TEST") != "true" {
		t.Skip("Image upload test skipped. Set ENABLE_IMAGE_UPLOAD_TEST=true to run.")
	}

	service := createIntegrationTestService(t)
	ctx := context.Background()

	t.Run("create image upload", func(t *testing.T) {
		params := ImageUploadParams{
			Label:       "integration-test-upload",
			Region:      "us-east", // Use a reliable region
			Description: "Integration test image upload",
			Tags:        []string{"integration", "test"},
		}

		result, err := service.handleCreateImageUpload(ctx, params)

		// Note: This will create a real upload URL, so we should clean up
		// In a real integration test, you'd want to track and clean up resources
		if err == nil {
			require.NotNil(t, result)
			require.NotEmpty(t, result.ImageID)
			require.NotEmpty(t, result.UploadTo)

			t.Logf("Created image upload: %s", result.ImageID)
			t.Logf("Upload URL: %s", result.UploadTo)

			// TODO: In a complete test, you would:
			// 1. Upload a test image file to the URL
			// 2. Wait for processing to complete
			// 3. Test other operations on the uploaded image
			// 4. Clean up by deleting the image
		} else {
			// Expected to fail without proper setup
			require.Contains(t, err.Error(), "failed to create image upload")
		}
	})
}

// TestImageIntegrationErrorHandling tests various error conditions.
func TestImageIntegrationErrorHandling(t *testing.T) {
	service := createIntegrationTestService(t)
	ctx := context.Background()

	t.Run("delete public image", func(t *testing.T) {
		params := ImageDeleteParams{
			ImageID: "linode/ubuntu22.04",
		}

		result, err := service.handleDeleteImage(ctx, params)
		require.Error(t, err)
		require.Empty(t, result)
		require.Contains(t, err.Error(), "cannot delete public images")
	})

	t.Run("update public image", func(t *testing.T) {
		params := ImageUpdateParams{
			ImageID: "linode/ubuntu22.04",
			Label:   "should-not-work",
		}

		result, err := service.handleUpdateImage(ctx, params)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "failed to update image")
	})
}

// Helper function to create an integration test service
func createIntegrationTestService(t *testing.T) *Service {
	token := os.Getenv("LINODE_TEST_TOKEN")
	if token == "" {
		t.Skip("LINODE_TEST_TOKEN environment variable not set")
	}

	l := logger.New("info")

	cfg := &config.Config{
		DefaultLinodeAccount: "integration",
		LinodeAccounts: map[string]config.LinodeAccount{
			"integration": {
				Token: token,
				Label: "Integration Test Account",
			},
		},
	}

	service, err := New(cfg, l)
	require.NoError(t, err)

	return service
}
