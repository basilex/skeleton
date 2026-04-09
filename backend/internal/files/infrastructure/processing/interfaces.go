package processing

import (
	"context"
)

// ImageMetadata contains metadata about an image.
type ImageMetadata struct {
	Width  int    // Image width in pixels
	Height int    // Image height in pixels
	Format string // Image format (jpeg, png, gif, etc.)
	Size   int64  // File size in bytes
}

// ImageProcessor defines the interface for image processing operations.
type ImageProcessor interface {
	// Resize resizes an image to the specified dimensions.
	// If width or height is 0, the aspect ratio is preserved.
	Resize(ctx context.Context, input []byte, width, height int) ([]byte, error)

	// Crop crops an image to the specified dimensions at the specified position.
	Crop(ctx context.Context, input []byte, x, y, width, height int) ([]byte, error)

	// Compress compresses an image with the specified quality (1-100).
	Compress(ctx context.Context, input []byte, quality int) ([]byte, error)

	// Convert converts an image to a different format.
	Convert(ctx context.Context, input []byte, format string) ([]byte, error)

	// GenerateThumbnail generates a thumbnail with the specified dimensions.
	GenerateThumbnail(ctx context.Context, input []byte, width, height int) ([]byte, error)

	// GetMetadata returns metadata about the image.
	GetMetadata(ctx context.Context, input []byte) (*ImageMetadata, error)

	// Watermark adds a watermark to the image.
	Watermark(ctx context.Context, input []byte, watermark string, position WatermarkPosition) ([]byte, error)
}

// WatermarkPosition specifies where to place the watermark.
type WatermarkPosition string

const (
	WatermarkTopLeft     WatermarkPosition = "top_left"
	WatermarkTopRight    WatermarkPosition = "top_right"
	WatermarkBottomLeft  WatermarkPosition = "bottom_left"
	WatermarkBottomRight WatermarkPosition = "bottom_right"
	WatermarkCenter      WatermarkPosition = "center"
)
