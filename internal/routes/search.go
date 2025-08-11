package routes

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rm-hull/street-manager-relay/internal"
	"github.com/rm-hull/street-manager-relay/models"
)

func HandleSearch(repo *internal.DbRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		bbox, err := models.BoundingBoxFromCSV(c.Query("bbox"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		facets, err := bindFacets(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Malformed facets"})
			return
		}

		events, err := repo.Search(bbox, facets)
		if err != nil {
			log.Printf("Error searching events: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search events"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"results":     events,
			"attribution": internal.ATTRIBUTION,
		})
	}
}

func bindFacets(c *gin.Context) (*models.Facets, error) {
	facets := &models.Facets{}

	// Bind string array parameters
	// Supports multiple formats:
	// 1. Multiple params: ?traffic_management_type_ref=Excavation&traffic_management_type_ref=Resurfacing
	// 2. Comma-separated: ?traffic_management_type_ref=Excavation,Resurfacing
	// 3. Empty means no filter

	if values := c.QueryArray("permit_status"); len(values) > 0 {
		facets.PermitStatus = expandCommaSeparated(values)
	}

	if values := c.QueryArray("traffic_management_type_ref"); len(values) > 0 {
		facets.TrafficManagementTypeRef = expandCommaSeparated(values)
	}

	if values := c.QueryArray("work_status_ref"); len(values) > 0 {
		facets.WorkStatusRef = expandCommaSeparated(values)
	}

	if values := c.QueryArray("work_category_ref"); len(values) > 0 {
		facets.WorkCategoryRef = expandCommaSeparated(values)
	}

	if values := c.QueryArray("road_category"); len(values) > 0 {
		facets.RoadCategory = expandCommaSeparated(values)
	}

	if values := c.QueryArray("highway_authority"); len(values) > 0 {
		facets.HighwayAuthority = expandCommaSeparated(values)
	}

	if values := c.QueryArray("promoter_organisation"); len(values) > 0 {
		facets.PromoterOrganisation = expandCommaSeparated(values)
	}

	return facets, nil
}

// expandCommaSeparated handles both multiple query params and comma-separated values
// e.g., both "?type=A&type=B" and "?type=A,B" result in []string{"A", "B"}
func expandCommaSeparated(values []string) []string {
	var result []string
	for _, value := range values {
		// Split by comma and trim whitespace
		parts := strings.Split(value, ",")
		for _, part := range parts {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				result = append(result, trimmed)
			}
		}
	}
	return result
}
