package image

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"log/slog"

	"github.com/aiservice/internal/models"
)

// ImageCropper handles image cropping with element-aware logic
type ImageCropper struct{}

// NewImageCropper creates a new image cropper
func NewImageCropper() *ImageCropper {
	return &ImageCropper{}
}

// CropWithElements crops an image ensuring specified elements are fully visible
func (c *ImageCropper) CropWithElements(
	imageData []byte,
	initialBBox models.BoundingBox,
	elements []models.Element,
) ([]byte, models.BoundingBox, error) {
	slog.Info("[ImageCropper.CropWithElements] Starting crop",
		"initialBBox", fmt.Sprintf("[%f,%f,%f,%f]", initialBBox.X, initialBBox.Y, initialBBox.W, initialBBox.H),
		"elementCount", len(elements))

	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		slog.Error("[ImageCropper.CropWithElements] Failed to decode image", "err", err)
		return nil, models.BoundingBox{}, fmt.Errorf("failed to decode image: %w", err)
	}

	slog.Info("[ImageCropper.CropWithElements] Image decoded",
		"imageWidth", img.Bounds().Dx(), "imageHeight", img.Bounds().Dy())

	// Expand bbox to include full elements
	expandedBBox := c.expandBBoxForElements(initialBBox, elements)

	slog.Info("[ImageCropper.CropWithElements] BBox expanded",
		"expandedBBox", fmt.Sprintf("[%f,%f,%f,%f]", expandedBBox.X, expandedBBox.Y, expandedBBox.W, expandedBBox.H))

	// Ensure bbox is within image bounds
	imgBounds := img.Bounds()
	expandedBBox = c.clampBBoxToImage(expandedBBox, imgBounds)

	slog.Info("[ImageCropper.CropWithElements] BBox clamped to image bounds",
		"clampedBBox", fmt.Sprintf("[%f,%f,%f,%f]", expandedBBox.X, expandedBBox.Y, expandedBBox.W, expandedBBox.H))

	// Crop the image
	croppedImg := c.cropImage(img, expandedBBox)

	// Encode back to bytes
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, croppedImg, &jpeg.Options{Quality: 85})
	if err != nil {
		slog.Error("[ImageCropper.CropWithElements] Failed to encode cropped image", "err", err)
		return nil, models.BoundingBox{}, fmt.Errorf("failed to encode cropped image: %w", err)
	}

	slog.Info("[ImageCropper.CropWithElements] Crop completed successfully",
		"croppedSize", buf.Len())

	return buf.Bytes(), expandedBBox, nil
}

// MergeOverlappingCrops merges overlapping bounding boxes
func (c *ImageCropper) MergeOverlappingCrops(bboxes []models.BoundingBox) []models.BoundingBox {
	if len(bboxes) == 0 {
		return bboxes
	}

	// Simple merge algorithm: repeatedly merge overlapping boxes
	merged := make([]models.BoundingBox, len(bboxes))
	copy(merged, bboxes)

	changed := true
	for changed {
		changed = false
		var newMerged []models.BoundingBox
		used := make([]bool, len(merged))

		for i := 0; i < len(merged); i++ {
			if used[i] {
				continue
			}

			current := merged[i]
			used[i] = true

			for j := i + 1; j < len(merged); j++ {
				if used[j] {
					continue
				}

				if current.Intersects(merged[j]) {
					current = current.Union(merged[j])
					used[j] = true
					changed = true
				}
			}

			newMerged = append(newMerged, current)
		}

		merged = newMerged
	}

	return merged
}

// ImageToBase64 converts image bytes to base64 string
func (c *ImageCropper) ImageToBase64(imageData []byte) string {
	return base64.StdEncoding.EncodeToString(imageData)
}

// ShouldFallbackToFullScan checks if incremental analysis should fallback to full scan
func (c *ImageCropper) ShouldFallbackToFullScan(crops []models.BoundingBox, boardSize models.BoundingBox) bool {
	slog.Info("[ImageCropper.ShouldFallbackToFullScan] Checking fallback condition",
		"cropCount", len(crops),
		"boardSize", fmt.Sprintf("[%f,%f]", boardSize.W, boardSize.H))

	if len(crops) == 0 {
		slog.Info("[ImageCropper.ShouldFallbackToFullScan] No crops, no fallback needed")
		return false
	}

	totalCropArea := 0.0
	for i, crop := range crops {
		area := crop.Area()
		totalCropArea += area
		slog.Info("[ImageCropper.ShouldFallbackToFullScan] Crop area",
			"cropIndex", i, "area", area, "totalArea", totalCropArea)
	}

	boardArea := boardSize.Area()
	if boardArea == 0 {
		slog.Warn("[ImageCropper.ShouldFallbackToFullScan] Board area is zero, forcing fallback")
		return true
	}

	ratio := totalCropArea / boardArea
	slog.Info("[ImageCropper.ShouldFallbackToFullScan] Fallback ratio calculation",
		"totalCropArea", totalCropArea,
		"boardArea", boardArea,
		"ratio", ratio,
		"threshold", 0.5)

	// Fallback if crops cover more than 50% of the board (dynamic threshold)
	// This is more conservative - if half the board is modified, full scan is often cheaper
	needsFallback := ratio > 0.5

	if needsFallback {
		slog.Info("[ImageCropper.ShouldFallbackToFullScan] Fallback needed", "ratio", ratio)
	} else {
		slog.Info("[ImageCropper.ShouldFallbackToFullScan] No fallback needed", "ratio", ratio)
	}

	return needsFallback
}

// expandBBoxForElements expands a bounding box to fully contain intersecting elements
func (c *ImageCropper) expandBBoxForElements(bbox models.BoundingBox, elements []models.Element) models.BoundingBox {
	expanded := bbox

	for _, elem := range elements {
		elemBBox := models.BoundingBox{
			X: elem.X,
			Y: elem.Y,
			W: elem.Width,
			H: elem.Height,
		}

		// If element intersects with bbox, expand to include it fully
		if bbox.Intersects(elemBBox) {
			expanded = expanded.Union(elemBBox)
		}
	}

	return expanded
}

// clampBBoxToImage ensures bounding box is within image bounds
func (c *ImageCropper) clampBBoxToImage(bbox models.BoundingBox, imgBounds image.Rectangle) models.BoundingBox {
	clamped := bbox

	if clamped.X < 0 {
		clamped.X = 0
	}
	if clamped.Y < 0 {
		clamped.Y = 0
	}
	if clamped.X+clamped.W > float32(imgBounds.Dx()) {
		clamped.W = float32(imgBounds.Dx()) - clamped.X
	}
	if clamped.Y+clamped.H > float32(imgBounds.Dy()) {
		clamped.H = float32(imgBounds.Dy()) - clamped.Y
	}

	return clamped
}

// cropImage crops an image to the specified bounding box
func (c *ImageCropper) cropImage(img image.Image, bbox models.BoundingBox) image.Image {
	rect := image.Rect(
		int(bbox.X),
		int(bbox.Y),
		int(bbox.X+bbox.W),
		int(bbox.Y+bbox.H),
	)

	cropped := image.NewRGBA(rect)
	draw.Draw(cropped, rect, img, rect.Min, draw.Src)

	return cropped
}
