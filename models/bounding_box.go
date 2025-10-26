package models

import (
	"math"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/twpayne/go-geom/encoding/wkt"
)

type BBox struct {
	MinX float64
	MaxX float64
	MinY float64
	MaxY float64
}

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

func (bbox BBox) Equals(other BBox, tolerance float64) bool {
	return almostEqual(bbox.MinX, other.MinX, tolerance) &&
		almostEqual(bbox.MaxX, other.MaxX, tolerance) &&
		almostEqual(bbox.MinY, other.MinY, tolerance) &&
		almostEqual(bbox.MaxY, other.MaxY, tolerance)
}

func BoundingBoxFromWKT(wktStr string) (*BBox, error) {
	g, err := wkt.Unmarshal(wktStr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse WKT")
	}

	bounds := g.Bounds()
	return &BBox{
		MinX: bounds.Min(0),
		MaxX: bounds.Max(0),
		MinY: bounds.Min(1),
		MaxY: bounds.Max(1),
	}, nil
}

func BoundingBoxFromCSV(bboxStr string) (*BBox, error) {
	bboxParts := strings.Split(bboxStr, ",")
	if len(bboxParts) != 4 {
		return nil, errors.New("bbox must have 4 comma-separated values")
	}

	bbox := make([]float64, 4)
	for i, part := range bboxParts {
		val, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil {
			return nil, errors.Newf("invalid bbox value '%s': not a valid float", part)
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
