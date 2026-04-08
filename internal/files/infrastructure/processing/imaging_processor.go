package processing

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
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
// The watermark parameter can be either:
// - A text string (for text watermarks)
// - A file path to a PNG image (for image watermarks)
func (p *ImagingProcessor) Watermark(ctx context.Context, input []byte, watermark string, position WatermarkPosition) ([]byte, error) {
	img, format, err := p.decodeImage(input)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	// Check if watermark is a file path (PNG or JPEG)
	if isImageFile(watermark) {
		return p.addImageWatermark(img, format, watermark, position)
	}

	// Otherwise, treat as text watermark
	return p.addTextWatermark(img, format, watermark, position)
}

// addTextWatermark adds a text watermark to the image.
func (p *ImagingProcessor) addTextWatermark(img image.Image, format, text string, position WatermarkPosition) ([]byte, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create a new RGBA image
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// Use basic font face
	face := basicfont.Face7x13

	// Calculate text dimensions
	textWidth := font.MeasureString(face, text).Ceil()
	textHeight := face.Metrics().Height.Ceil()

	// Calculate watermark position with padding
	padding := 10
	x, y := calculatePosition(position, width, height, textWidth, textHeight, padding)

	// Create font drawer for shadow (semi-transparent black)
	point := fixed.Point26_6{
		X: fixed.I(x + 1),
		Y: fixed.I(y + textHeight + 1),
	}

	d := &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(color.RGBA{0, 0, 0, 128}),
		Face: face,
		Dot:  point,
	}
	d.DrawString(text)

	// Create font drawer for main text (semi-transparent white)
	point.X = fixed.I(x)
	point.Y = fixed.I(y + textHeight)

	d = &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(color.RGBA{255, 255, 255, 200}),
		Face: face,
		Dot:  point,
	}
	d.DrawString(text)

	return p.encodeImage(rgba, format)
}

// addImageWatermark adds an image watermark (PNG) to the image.
func (p *ImagingProcessor) addImageWatermark(img image.Image, format, watermarkPath string, position WatermarkPosition) ([]byte, error) {
	// Read watermark image
	watermarkBytes, err := os.ReadFile(watermarkPath)
	if err != nil {
		return nil, fmt.Errorf("read watermark file: %w", err)
	}

	// Decode watermark image
	watermarkImg, _, err := p.decodeImage(watermarkBytes)
	if err != nil {
		return nil, fmt.Errorf("decode watermark image: %w", err)
	}

	// Get dimensions
	imgBounds := img.Bounds()
	watermarkBounds := watermarkImg.Bounds()

	// Calculate position
	x, y := calculatePosition(position, imgBounds.Dx(), imgBounds.Dy(), watermarkBounds.Dx(), watermarkBounds.Dy(), 10)

	// Overlay watermark on the image
	result := imaging.Overlay(img, watermarkImg, image.Pt(x, y), 0.7)

	return p.encodeImage(result, format)
}

// calculatePosition calculates the x, y coordinates for the watermark based on position.
func calculatePosition(position WatermarkPosition, imgWidth, imgHeight, watermarkWidth, watermarkHeight, padding int) (int, int) {
	switch position {
	case WatermarkTopLeft:
		return padding, padding
	case WatermarkTopRight:
		return imgWidth - watermarkWidth - padding, padding
	case WatermarkBottomLeft:
		return padding, imgHeight - watermarkHeight - padding
	case WatermarkBottomRight:
		return imgWidth - watermarkWidth - padding, imgHeight - watermarkHeight - padding
	case WatermarkCenter:
		return (imgWidth - watermarkWidth) / 2, (imgHeight - watermarkHeight) / 2
	default:
		return padding, padding
	}
}

// isImageFile checks if the watermark string is a path to an image file.
func isImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg"
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
