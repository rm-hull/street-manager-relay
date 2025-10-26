package routes

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/rm-hull/street-manager-relay/internal"
	"github.com/rm-hull/street-manager-relay/internal/promoter"
	"github.com/rm-hull/street-manager-relay/models"
)

func HandleSearch(repo *internal.DbRepository, organisations promoter.Organisations) gin.HandlerFunc {
	return func(c *gin.Context) {
		bbox, err := models.BoundingBoxFromCSV(c.Query("bbox"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		facets, err := bindFacets(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Malformed facets"})
			return
		}

		temporalFilters, err := bindTemporalFilters(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		events, err := repo.Search(bbox, facets, temporalFilters)
		if err != nil {
			_ = c.Error(errors.Wrap(err, "error searching events"))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to search events"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"results":     enrich(organisations, events),
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

func bindTemporalFilters(c *gin.Context) (*models.TemporalFilters, error) {
	filters := models.TemporalFilters{
		MaxDaysAhead:  7,
		MaxDaysBehind: 0,
	}

	params := map[string]*int{
		"max_days_ahead":  &filters.MaxDaysAhead,
		"max_days_behind": &filters.MaxDaysBehind,
	}

	for key, target := range params {
		if value := c.Query(key); value != "" {
			num, err := strconv.Atoi(value)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to convert %s", key)
			}
			if num < 0 {
				return nil, errors.Newf("%s must be non-negative, but got %d", key, num)
			}
			*target = num
		}
	}

	return &filters, nil
}

type EnrichedEvent struct {
	*models.Event
	PromoterWebsiteURL *string `json:"promoter_website_url,omitempty"`
	PromoterLogoURL    *string `json:"promoter_logo_url,omitempty"`
}

func enrich(promoterOrgs promoter.Organisations, events []*models.Event) []*EnrichedEvent {
	out := make([]*EnrichedEvent, len(events))

	for idx, event := range events {

		enrichedEvent := &EnrichedEvent{Event: event}

		if event.PromoterSWACode != nil {
			if org, ok := promoterOrgs[*event.PromoterSWACode]; ok {
				enrichedEvent.PromoterWebsiteURL = &org.Url
				enrichedEvent.PromoterLogoURL = org.Favicon
			}
		}
		out[idx] = enrichedEvent
	}
	return out
}
