package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rm-hull/street-manager-relay/generated"
	"github.com/rm-hull/street-manager-relay/internal"
	"github.com/rm-hull/street-manager-relay/models"
)

func HandleSNSMessage(repo *internal.DbRepository, certManager internal.CertManager) gin.HandlerFunc {
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
