package processing

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/disintegration/imaging"
)

// ImagingProcessor implements ImageProcessor using disintegration/imaging.
type ImagingProcessor struct{}

// NewImagingProcessor creates a new imaging processor.
func NewImagingProcessor() *ImagingProcessor {
	return &ImagingProcessor{}
}

// Resize resizes an image to the specified dimensions.
func (p *ImagingProcessor) Resize(ctx context.Context, input []byte, width, height int) ([]byte, error) {
	img, format, err := p.decodeImage(input)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	// Resize preserving aspect ratio if one dimension is 0
	var resized *image.NRGBA
	if width == 0 && height > 0 {
		resized = imaging.Resize(img, 0, height, imaging.Lanczos)
	} else if height == 0 && width > 0 {
		resized = imaging.Resize(img, width, 0, imaging.Lanczos)
	} else {
		resized = imaging.Resize(img, width, height, imaging.Lanczos)
	}

	return p.encodeImage(resized, format)
}

// Crop crops an image to the specified dimensions at the specified position.
func (p *ImagingProcessor) Crop(ctx context.Context, input []byte, x, y, width, height int) ([]byte, error) {
	img, format, err := p.decodeImage(input)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	cropped := imaging.Crop(img, image.Rect(x, y, x+width, y+height))

	return p.encodeImage(cropped, format)
}

// Compress compresses an image with the specified quality (1-100).
func (p *ImagingProcessor) Compress(ctx context.Context, input []byte, quality int) ([]byte, error) {
	if quality < 1 || quality > 100 {
		return nil, fmt.Errorf("quality must be between 1 and 100")
	}

	img, format, err := p.decodeImage(input)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	// Convert to JPEG for compression
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, fmt.Errorf("encode jpeg: %w", err)
	}

	// If original was PNG and we want to keep it as PNG, re-encode
	if format == "png" {
		buf.Reset()
		if err := png.Encode(buf, img); err != nil {
			return nil, fmt.Errorf("encode png: %w", err)
		}
	}

	return buf.Bytes(), nil
}

// Convert converts an image to a different format.
func (p *ImagingProcessor) Convert(ctx context.Context, input []byte, format string) ([]byte, error) {
	img, _, err := p.decodeImage(input)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	buf := new(bytes.Buffer)

	switch format {
	case "jpeg", "jpg":
		if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 85}); err != nil {
			return nil, fmt.Errorf("encode jpeg: %w", err)
		}
	case "png":
		if err := png.Encode(buf, img); err != nil {
			return nil, fmt.Errorf("encode png: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	return buf.Bytes(), nil
}

// GenerateThumbnail generates a thumbnail with the specified dimensions.
func (p *ImagingProcessor) GenerateThumbnail(ctx context.Context, input []byte, width, height int) ([]byte, error) {
	img, _, err := p.decodeImage(input)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	// Generate thumbnail preserving aspect ratio
	thumbnail := imaging.Thumbnail(img, width, height, imaging.Lanczos)

	// Encode as JPEG
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, thumbnail, &jpeg.Options{Quality: 85}); err != nil {
		return nil, fmt.Errorf("encode thumbnail: %w", err)
	}

	return buf.Bytes(), nil
}

// GetMetadata returns metadata about the image.
func (p *ImagingProcessor) GetMetadata(ctx context.Context, input []byte) (*ImageMetadata, error) {
	img, format, err := p.decodeImage(input)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	bounds := img.Bounds()

	return &ImageMetadata{
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
		Format: format,
		Size:   int64(len(input)),
	}, nil
}

// Watermark adds a watermark to the image.
// Note: This is a simplified implementation. For production use,
// consider using a more sophisticated watermarking library.
func (p *ImagingProcessor) Watermark(ctx context.Context, input []byte, watermark string, position WatermarkPosition) ([]byte, error) {
	img, format, err := p.decodeImage(input)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	// For now, we'll just return the original image
	// TODO: Implement proper watermarking with text or image overlay
	// This would typically use golang.org/x/image/font or similar

	return p.encodeImage(img, format)
}

// decodeImage decodes an image from bytes and returns the image and format.
func (p *ImagingProcessor) decodeImage(input []byte) (image.Image, string, error) {
	reader := bytes.NewReader(input)

	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, "", fmt.Errorf("decode: %w", err)
	}

	return img, format, nil
}

// encodeImage encodes an image to bytes in the specified format.
func (p *ImagingProcessor) encodeImage(img image.Image, format string) ([]byte, error) {
	buf := new(bytes.Buffer)

	switch format {
	case "jpeg", "jpg":
		if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 85}); err != nil {
			return nil, fmt.Errorf("encode jpeg: %w", err)
		}
	case "png":
		if err := png.Encode(buf, img); err != nil {
			return nil, fmt.Errorf("encode png: %w", err)
		}
	default:
		// Default to JPEG
		if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 85}); err != nil {
			return nil, fmt.Errorf("encode jpeg: %w", err)
		}
	}

	return buf.Bytes(), nil
}
