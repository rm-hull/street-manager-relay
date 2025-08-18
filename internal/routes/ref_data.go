package routes

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/kofalt/go-memoize"
	"github.com/rm-hull/street-manager-relay/internal"
	"github.com/rm-hull/street-manager-relay/models"
)

func HandleRefData(repo *internal.DbRepository, cache *memoize.Memoizer) gin.HandlerFunc {
	return func(c *gin.Context) {

		refData, err, _ := memoize.Call(cache, "refdata", func() (*models.RefData, error) {
			return repo.RefData()
		})
		if err != nil {
			log.Printf("Error fetching reference data: %v", err)
			c.JSON(500, gin.H{"error": "Failed to fetch reference data"})
			return
		}

		c.JSON(200, gin.H{
			"refdata":     refData,
			"attribution": internal.ATTRIBUTION,
		})
	}
}
