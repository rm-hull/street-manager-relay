package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/kofalt/go-memoize"
	"github.com/rm-hull/street-manager-relay/internal"
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
	)

	err := healthcheck.New(r, hc_config.DefaultConfig(), []checks.Check{})
	if err != nil {
		log.Fatalf("failed to initialize healthcheck: %v", err)
	}

	cache := memoize.NewMemoizer(24*time.Hour, 1*time.Hour)
	certManager := internal.NewCertManager(cache)

	r.POST("/v1/street-manager-relay/sns", handleSNSMessage(dataPath, certManager))

	log.Printf("HTTP subscriber listening at http://localhost:%d", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func handleSNSMessage(dataPath string, certManager internal.CertManager) func(c *gin.Context) {
	log.Printf("Storing incoming messages to path: %s", dataPath)
	if err := os.MkdirAll(dataPath, os.ModePerm); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

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

		if err := writeTimestampedFile(dataPath, bodyBytes); err != nil {
			log.Printf("Error writing file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write file"})
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

		if err := handleMessage(&body); err != nil {
			log.Printf("Error handling message: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to handle message"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

func handleMessage(body *internal.SNSMessage) error {
	switch body.Type {
	case "SubscriptionConfirmation":
		return confirmSubscription(body.SubscribeURL)
	case "Notification":
		return handleNotification(body)
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

func handleNotification(body *internal.SNSMessage) error {
	log.Printf("Received message from SNS: %s", body.Message)
	return nil
}

func writeTimestampedFile(dataPath string, bodyBytes []byte) error {
	today := time.Now()
	dateDir := today.Format("2006-01-02") // YYYY-MM-DD format

	fullDirPath := filepath.Join(dataPath, dateDir)

	// Create the directory if it doesn't exist.
	if err := os.MkdirAll(fullDirPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	epochMillis := time.Now().UnixMilli()
	fileName := fmt.Sprintf("%05d.%03d.json", (epochMillis/1000)%86_400, epochMillis%1000)
	filePath := filepath.Join(fullDirPath, fileName)

	if err := os.WriteFile(filePath, bodyBytes, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
