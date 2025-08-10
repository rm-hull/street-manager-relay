package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Depado/ginprom"
	"github.com/aurowora/compress"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/kofalt/go-memoize"
	"github.com/rm-hull/street-manager-relay/generated"
	"github.com/rm-hull/street-manager-relay/internal"
	"github.com/rm-hull/street-manager-relay/models"
	"github.com/tavsec/gin-healthcheck/checks"

	healthcheck "github.com/tavsec/gin-healthcheck"
	hc_config "github.com/tavsec/gin-healthcheck/config"
)

func ApiServer(dbPath string, port int, debug bool) {

	repo, err := internal.NewDbRepository(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize db repository: %v", err)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	r := gin.New()

	prometheus := ginprom.New(
		ginprom.Engine(r),
		ginprom.Path("/metrics"),
		ginprom.Ignore("/healthz"),
	)

	r.Use(
		gin.Recovery(),
		gin.LoggerWithWriter(gin.DefaultWriter, "/healthz", "/metrics"),
		prometheus.Instrument(),
		compress.Compress(),
		cors.Default(),
	)

	if debug {
		log.Println("WARNING: pprof endpoints are enabled and exposed. Do not run with this flag in production.")
		pprof.Register(r)
	}

	err = healthcheck.New(r, hc_config.DefaultConfig(), []checks.Check{
		repo.HealthCheck(),
	})
	if err != nil {
		log.Fatalf("failed to initialize healthcheck: %v", err)
	}

	cache := memoize.NewMemoizer(24*time.Hour, 1*time.Hour)
	certManager := internal.NewCertManager(cache)

	r.POST("/v1/street-manager-relay/sns", handleSNSMessage(repo, certManager))
	r.GET("/v1/street-manager-relay/search", handleSearch(repo))

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting HTTP API Server on port %d...", port)
	err = r.Run(addr)
	log.Fatalf("HTTP API Server failed to start on port %d: %v", port, err)
}

func handleSearch(repo *internal.DbRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		bbox, err := models.BoundingBoxFromCSV(c.Query("bbox"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		events, err := repo.Search(bbox)
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

func handleSNSMessage(repo *internal.DbRepository, certManager internal.CertManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		messageType := c.GetHeader("x-amz-sns-message-type")
		if messageType == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing x-amz-sns-message-type header"})
			return
		}

		bodyBytes, err := c.GetRawData()
		if err != nil {
			log.Printf("Error reading request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}

		var body internal.SNSMessage
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			log.Printf("Error parsing JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		valid, err := internal.IsValidSignature(&body, certManager)
		if err != nil {
			log.Printf("Error validating signature: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Signature validation failed"})
			return
		}

		if !valid {
			log.Println("Message signature is not valid")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Message signature is not valid"})
			return
		}

		if err := handleMessage(repo, &body); err != nil {
			log.Printf("Error handling message: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to handle message"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

func handleMessage(repo *internal.DbRepository, body *internal.SNSMessage) error {
	switch body.Type {
	case "SubscriptionConfirmation":
		return confirmSubscription(body.SubscribeURL)
	case "Notification":
		return handleNotification(repo, body)
	default:
		log.Printf("Unknown message type: %s", body.Type)
		return nil
	}
}

// confirmSubscription confirms the SNS subscription by making GET request to subscribe URL
func confirmSubscription(subscriptionURL string) error {
	resp, err := http.Get(subscriptionURL)
	if err != nil {
		return fmt.Errorf("failed to confirm subscription: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("subscription confirmation failed with status: %d", resp.StatusCode)
	}

	log.Println("Subscription confirmed")
	return nil
}

func handleNotification(repo *internal.DbRepository, body *internal.SNSMessage) error {
	event, err := generated.UnmarshalEventNotifierMessage([]byte(body.Message))
	if err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	batch, err := repo.BatchUpsert()
	if err != nil {
		return fmt.Errorf("failed to create batch upserter: %w", err)
	}

	if _, err = batch.Upsert(models.NewEventFrom(event)); err != nil {
		return fmt.Errorf("failed to upsert: %w", batch.Abort(err))
	}

	return batch.Done()
}
