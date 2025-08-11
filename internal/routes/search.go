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
	binders := map[string]func([]string){
		"permit_status":               func(v []string) { facets.PermitStatus = v },
		"traffic_management_type_ref": func(v []string) { facets.TrafficManagementTypeRef = v },
		"work_status_ref":             func(v []string) { facets.WorkStatusRef = v },
		"work_category_ref":           func(v []string) { facets.WorkCategoryRef = v },
		"road_category":               func(v []string) { facets.RoadCategory = v },
		"highway_authority":           func(v []string) { facets.HighwayAuthority = v },
		"promoter_organisation":       func(v []string) { facets.PromoterOrganisation = v },
	}

	for param, setter := range binders {
		if values := c.QueryArray(param); len(values) > 0 {
			setter(expandCommaSeparated(values))
		}
	}

	return facets, nil
}

// expandCommaSeparated handles both multiple query params and comma-separated values
// e.g., both "?type=A&type=B" and "?type=A,B" result in []string{"A", "B"}
func expandCommaSeparated(values []string) []string {
	var result []string
	for _, value := range values {
		parts := strings.SplitSeq(value, ",")
		for part := range parts {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				result = append(result, trimmed)
			}
		}
	}
	return result
}
