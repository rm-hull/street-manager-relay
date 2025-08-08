package models

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/twpayne/go-geom/encoding/wkt"
)

type BBox struct {
	MinX float64
	MaxX float64
	MinY float64
	MaxY float64
}

func BoundingBoxFromWKT(wktStr string) (*BBox, error) {
	g, err := wkt.Unmarshal(wktStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse WKT: %w", err)
	}

	coords := g.FlatCoords()
	if len(coords) < 2 {
		return nil, fmt.Errorf("invalid coordinates")
	}

	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64

	for i := 0; i < len(coords); i += 2 {
		x := coords[i]
		y := coords[i+1]

		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}
		if x > maxX {
			maxX = x
		}
		if y > maxY {
			maxY = y
		}
	}

	return &BBox{MinX: minX, MaxX: maxX, MinY: minY, MaxY: maxY}, nil
}

func BoundingBoxFromCSV(bboxStr string) (*BBox, error) {
	bboxParts := strings.Split(bboxStr, ",")
	if len(bboxParts) != 4 {
		return nil, fmt.Errorf("bbox must have 4 comma-separated values")
	}

	bbox := make([]float64, 4)
	for i, part := range bboxParts {
		val, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid bbox value '%s': not a valid float", part)
		}
		bbox[i] = val
	}

	return &BBox{
		MinX: math.Min(bbox[0], bbox[2]),
		MaxX: math.Max(bbox[0], bbox[2]),
		MinY: math.Min(bbox[1], bbox[3]),
		MaxY: math.Max(bbox[1], bbox[3]),
	}, nil
}
