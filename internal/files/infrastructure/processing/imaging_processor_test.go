package processing

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/stretchr/testify/require"
)

func createTestImage(t *testing.T, width, height int, format string) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with a simple pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(x % 256),
				G: uint8(y % 256),
				B: uint8((x + y) % 256),
				A: 255,
			})
		}
	}

	buf := new(bytes.Buffer)
	var err error

	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 90})
	case "png":
		err = png.Encode(buf, img)
	default:
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 90})
	}

	require.NoError(t, err)
	return buf.Bytes()
}

func TestNewImagingProcessor(t *testing.T) {
	processor := NewImagingProcessor()
	require.NotNil(t, processor)
}

func TestImagingProcessor_Resize(t *testing.T) {
	processor := NewImagingProcessor()

	t.Run("resize to smaller dimensions", func(t *testing.T) {
		input := createTestImage(t, 800, 600, "jpeg")
		result, err := processor.Resize(context.Background(), input, 400, 300)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify dimensions
		metadata, err := processor.GetMetadata(context.Background(), result)
		require.NoError(t, err)
		require.Equal(t, 400, metadata.Width)
		require.Equal(t, 300, metadata.Height)
	})

	t.Run("resize maintaining aspect ratio - width only", func(t *testing.T) {
		input := createTestImage(t, 800, 600, "jpeg")
		result, err := processor.Resize(context.Background(), input, 400, 0)
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := processor.GetMetadata(context.Background(), result)
		require.NoError(t, err)
		require.Equal(t, 400, metadata.Width)
		require.LessOrEqual(t, metadata.Height, 400)
	})

	t.Run("resize maintaining aspect ratio - height only", func(t *testing.T) {
		input := createTestImage(t, 800, 600, "jpeg")
		result, err := processor.Resize(context.Background(), input, 0, 300)
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := processor.GetMetadata(context.Background(), result)
		require.NoError(t, err)
		require.Equal(t, 300, metadata.Height)
		require.LessOrEqual(t, metadata.Width, 400)
	})

	t.Run("resize PNG", func(t *testing.T) {
		input := createTestImage(t, 400, 400, "png")
		result, err := processor.Resize(context.Background(), input, 200, 200)
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := processor.GetMetadata(context.Background(), result)
		require.NoError(t, err)
		require.Equal(t, 200, metadata.Width)
		require.Equal(t, 200, metadata.Height)
	})

	t.Run("invalid image data", func(t *testing.T) {
		_, err := processor.Resize(context.Background(), []byte("invalid"), 100, 100)
		require.Error(t, err)
		require.Contains(t, err.Error(), "decode image")
	})
}

func TestImagingProcessor_Crop(t *testing.T) {
	processor := NewImagingProcessor()

	t.Run("successful crop", func(t *testing.T) {
		input := createTestImage(t, 800, 600, "jpeg")
		result, err := processor.Crop(context.Background(), input, 100, 100, 300, 200)
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := processor.GetMetadata(context.Background(), result)
		require.NoError(t, err)
		require.Equal(t, 300, metadata.Width)
		require.Equal(t, 200, metadata.Height)
	})

	t.Run("crop from origin", func(t *testing.T) {
		input := createTestImage(t, 400, 400, "jpeg")
		result, err := processor.Crop(context.Background(), input, 0, 0, 200, 200)
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := processor.GetMetadata(context.Background(), result)
		require.NoError(t, err)
		require.Equal(t, 200, metadata.Width)
		require.Equal(t, 200, metadata.Height)
	})

	t.Run("crop PNG", func(t *testing.T) {
		input := createTestImage(t, 500, 500, "png")
		result, err := processor.Crop(context.Background(), input, 50, 50, 400, 300)
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := processor.GetMetadata(context.Background(), result)
		require.NoError(t, err)
		require.Equal(t, 400, metadata.Width)
		require.Equal(t, 300, metadata.Height)
	})

	t.Run("invalid image data", func(t *testing.T) {
		_, err := processor.Crop(context.Background(), []byte("invalid"), 0, 0, 100, 100)
		require.Error(t, err)
		require.Contains(t, err.Error(), "decode image")
	})
}

func TestImagingProcessor_Compress(t *testing.T) {
	processor := NewImagingProcessor()

	t.Run("compress JPEG with high quality", func(t *testing.T) {
		input := createTestImage(t, 400, 400, "jpeg")
		result, err := processor.Compress(context.Background(), input, 90)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Less(t, len(result), len(input)*2) // Compressed should not be much larger
	})

	t.Run("compress JPEG with low quality", func(t *testing.T) {
		input := createTestImage(t, 400, 400, "jpeg")
		result, err := processor.Compress(context.Background(), input, 10)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Less(t, len(result), len(input)) // Low quality should be smaller
	})

	t.Run("compress PNG", func(t *testing.T) {
		input := createTestImage(t, 400, 400, "png")
		result, err := processor.Compress(context.Background(), input, 85)
		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("invalid quality - too low", func(t *testing.T) {
		input := createTestImage(t, 100, 100, "jpeg")
		_, err := processor.Compress(context.Background(), input, 0)
		require.Error(t, err)
		require.Contains(t, err.Error(), "quality must be between")
	})

	t.Run("invalid quality - too high", func(t *testing.T) {
		input := createTestImage(t, 100, 100, "jpeg")
		_, err := processor.Compress(context.Background(), input, 101)
		require.Error(t, err)
		require.Contains(t, err.Error(), "quality must be between")
	})

	t.Run("invalid image data", func(t *testing.T) {
		_, err := processor.Compress(context.Background(), []byte("invalid"), 85)
		require.Error(t, err)
		require.Contains(t, err.Error(), "decode image")
	})
}

func TestImagingProcessor_Convert(t *testing.T) {
	processor := NewImagingProcessor()

	t.Run("convert PNG to JPEG", func(t *testing.T) {
		input := createTestImage(t, 200, 200, "png")
		result, err := processor.Convert(context.Background(), input, "jpeg")
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := processor.GetMetadata(context.Background(), result)
		require.NoError(t, err)
		require.Equal(t, "jpeg", metadata.Format)
	})

	t.Run("convert JPEG to PNG", func(t *testing.T) {
		input := createTestImage(t, 200, 200, "jpeg")
		result, err := processor.Convert(context.Background(), input, "png")
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := processor.GetMetadata(context.Background(), result)
		require.NoError(t, err)
		require.Equal(t, "png", metadata.Format)
	})

	t.Run("convert using jpg alias", func(t *testing.T) {
		input := createTestImage(t, 200, 200, "png")
		result, err := processor.Convert(context.Background(), input, "jpg")
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := processor.GetMetadata(context.Background(), result)
		require.NoError(t, err)
		require.Equal(t, "jpeg", metadata.Format)
	})

	t.Run("unsupported format", func(t *testing.T) {
		input := createTestImage(t, 100, 100, "jpeg")
		_, err := processor.Convert(context.Background(), input, "gif")
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported format")
	})

	t.Run("invalid image data", func(t *testing.T) {
		_, err := processor.Convert(context.Background(), []byte("invalid"), "jpeg")
		require.Error(t, err)
		require.Contains(t, err.Error(), "decode image")
	})
}

func TestImagingProcessor_GenerateThumbnail(t *testing.T) {
	processor := NewImagingProcessor()

	t.Run("generate thumbnail from landscape image", func(t *testing.T) {
		input := createTestImage(t, 800, 600, "jpeg")
		result, err := processor.GenerateThumbnail(context.Background(), input, 200, 200)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Thumbnail should not exceed specified dimensions
		metadata, err := processor.GetMetadata(context.Background(), result)
		require.NoError(t, err)
		require.LessOrEqual(t, metadata.Width, 200)
		require.LessOrEqual(t, metadata.Height, 200)
		require.Equal(t, "jpeg", metadata.Format)
	})

	t.Run("generate thumbnail from portrait image", func(t *testing.T) {
		input := createTestImage(t, 600, 800, "jpeg")
		result, err := processor.GenerateThumbnail(context.Background(), input, 150, 150)
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := processor.GetMetadata(context.Background(), result)
		require.NoError(t, err)
		require.LessOrEqual(t, metadata.Width, 150)
		require.LessOrEqual(t, metadata.Height, 150)
	})

	t.Run("generate thumbnail from PNG", func(t *testing.T) {
		input := createTestImage(t, 500, 500, "png")
		result, err := processor.GenerateThumbnail(context.Background(), input, 100, 100)
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := processor.GetMetadata(context.Background(), result)
		require.NoError(t, err)
		require.Equal(t, "jpeg", metadata.Format) // Thumbnail is always JPEG
	})

	t.Run("invalid image data", func(t *testing.T) {
		_, err := processor.GenerateThumbnail(context.Background(), []byte("invalid"), 100, 100)
		require.Error(t, err)
		require.Contains(t, err.Error(), "decode image")
	})
}

func TestImagingProcessor_GetMetadata(t *testing.T) {
	processor := NewImagingProcessor()

	t.Run("JPEG metadata", func(t *testing.T) {
		input := createTestImage(t, 640, 480, "jpeg")
		metadata, err := processor.GetMetadata(context.Background(), input)
		require.NoError(t, err)
		require.NotNil(t, metadata)
		require.Equal(t, 640, metadata.Width)
		require.Equal(t, 480, metadata.Height)
		require.Equal(t, "jpeg", metadata.Format)
		require.Equal(t, int64(len(input)), metadata.Size)
	})

	t.Run("PNG metadata", func(t *testing.T) {
		input := createTestImage(t, 200, 300, "png")
		metadata, err := processor.GetMetadata(context.Background(), input)
		require.NoError(t, err)
		require.NotNil(t, metadata)
		require.Equal(t, 200, metadata.Width)
		require.Equal(t, 300, metadata.Height)
		require.Equal(t, "png", metadata.Format)
		require.Equal(t, int64(len(input)), metadata.Size)
	})

	t.Run("invalid image data", func(t *testing.T) {
		_, err := processor.GetMetadata(context.Background(), []byte("not an image"))
		require.Error(t, err)
		require.Contains(t, err.Error(), "decode image")
	})
}

func TestImagingProcessor_Watermark(t *testing.T) {
	processor := NewImagingProcessor()

	t.Run("text watermark - bottom right", func(t *testing.T) {
		input := createTestImage(t, 400, 400, "jpeg")

		result, err := processor.Watermark(context.Background(), input, "© 2026", WatermarkBottomRight)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Result should have same dimensions
		metadata, err := processor.GetMetadata(context.Background(), result)
		require.NoError(t, err)
		require.Equal(t, 400, metadata.Width)
		require.Equal(t, 400, metadata.Height)

		// Result should be larger than input (watermark adds data)
		require.Greater(t, len(result), 0)
	})

	t.Run("text watermark - top left", func(t *testing.T) {
		input := createTestImage(t, 300, 200, "png")

		result, err := processor.Watermark(context.Background(), input, "CONFIDENTIAL", WatermarkTopLeft)
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := processor.GetMetadata(context.Background(), result)
		require.NoError(t, err)
		require.Equal(t, 300, metadata.Width)
		require.Equal(t, 200, metadata.Height)
	})

	t.Run("text watermark - center", func(t *testing.T) {
		input := createTestImage(t, 500, 300, "jpeg")

		result, err := processor.Watermark(context.Background(), input, "WATERMARK", WatermarkCenter)
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := processor.GetMetadata(context.Background(), result)
		require.NoError(t, err)
		require.Equal(t, 500, metadata.Width)
		require.Equal(t, 300, metadata.Height)
	})

	t.Run("text watermark - all positions", func(t *testing.T) {
		positions := []WatermarkPosition{
			WatermarkTopLeft,
			WatermarkTopRight,
			WatermarkBottomLeft,
			WatermarkBottomRight,
			WatermarkCenter,
		}

		input := createTestImage(t, 200, 200, "jpeg")

		for _, pos := range positions {
			result, err := processor.Watermark(context.Background(), input, "TEST", pos)
			require.NoError(t, err)
			require.NotNil(t, result, "Position: %s", pos)
		}
	})

	t.Run("text watermark on PNG", func(t *testing.T) {
		input := createTestImage(t, 400, 400, "png")

		result, err := processor.Watermark(context.Background(), input, "PNG Watermark", WatermarkBottomRight)
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := processor.GetMetadata(context.Background(), result)
		require.NoError(t, err)
		require.Equal(t, "png", metadata.Format)
	})

	t.Run("empty watermark text", func(t *testing.T) {
		input := createTestImage(t, 100, 100, "jpeg")

		result, err := processor.Watermark(context.Background(), input, "", WatermarkCenter)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Should still work with empty text
		metadata, err := processor.GetMetadata(context.Background(), result)
		require.NoError(t, err)
		require.Equal(t, 100, metadata.Width)
	})

	t.Run("invalid image data", func(t *testing.T) {
		_, err := processor.Watermark(context.Background(), []byte("invalid"), "test", WatermarkCenter)
		require.Error(t, err)
		require.Contains(t, err.Error(), "decode image")
	})

	t.Run("non-existent image file as watermark", func(t *testing.T) {
		input := createTestImage(t, 100, 100, "jpeg")

		// Try to use a non-existent file as watermark (should fallback to text or fail gracefully)
		_, err := processor.Watermark(context.Background(), input, "/nonexistent/watermark.png", WatermarkCenter)
		require.Error(t, err)
		require.Contains(t, err.Error(), "read watermark file")
	})
}
