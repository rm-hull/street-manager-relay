package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/aurowora/compress"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/kofalt/go-memoize"
	"github.com/rm-hull/street-manager-relay/generated"
	"github.com/rm-hull/street-manager-relay/internal"
	"github.com/rm-hull/street-manager-relay/models"
	"github.com/spf13/cobra"

	"github.com/tavsec/gin-healthcheck/checks"

	healthcheck "github.com/tavsec/gin-healthcheck"
	hc_config "github.com/tavsec/gin-healthcheck/config"
)

func main() {
	var err error
	var dataPath string
	var port int

	internal.ShowVersion()

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	internal.EnvironmentVars()

	rootCmd := &cobra.Command{
		Use:   "http",
		Short: "street manager relay server",
		Run: func(cmd *cobra.Command, args []string) {
			server(dataPath, port)
		},
	}

	rootCmd.Flags().StringVar(&dataPath, "data", "./data", "Path storaege for incoming messages")
	rootCmd.Flags().IntVar(&port, "port", 8080, "Port to run HTTP server on")

	if err = rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func server(dataPath string, port int) {
	r := gin.New()
	r.Use(
		gin.Recovery(),
		gin.LoggerWithWriter(gin.DefaultWriter, "/healthz"),
		compress.Compress(),
		cors.Default(),
	)

	err := healthcheck.New(r, hc_config.DefaultConfig(), []checks.Check{})
	if err != nil {
		log.Fatalf("failed to initialize healthcheck: %v", err)
	}

	cache := memoize.NewMemoizer(24*time.Hour, 1*time.Hour)
	certManager := internal.NewCertManager(cache)

	repo, err := internal.NewDbRepository(filepath.Join(dataPath, "street-manager.db"))
	if err != nil {
		log.Fatalf("Failed to initialize db repository: %v", err)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	r.POST("/v1/street-manager-relay/sns", handleSNSMessage(repo, certManager))
	r.GET("/v1/street-manager-relay/search", handleSearch(repo))

	log.Printf("HTTP subscriber listening at http://localhost:%d", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func handleSearch(repo *internal.DbRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		bbox, err := models.BoundingBoxFromCSV(c.Query("bbox"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		activities, err := repo.Search(bbox)
		if err != nil {
			log.Printf("Error searching activities: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search activities"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"activities": activities})
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

		// Only interested in activity topic messages (TODO: should really unsubscribe from others)
		if body.TopicArn != "arn:aws:sns:eu-west-2:287813576808:prod-activity-topic" {
			c.JSON(http.StatusAccepted, gin.H{"status": "ignored", "message": "Not processing notifications from this topic"})
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
	activity := models.NewActivityFrom(event)
	_, err = repo.Upsert(activity)

	return err
}
