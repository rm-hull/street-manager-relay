package models

import (
	"fmt"
	"math"

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
		return &BBox{}, fmt.Errorf("failed to parse WKT: %w", err)
	}

	coords := g.FlatCoords()
	if len(coords) < 2 {
		return &BBox{}, fmt.Errorf("invalid coordinates")
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
