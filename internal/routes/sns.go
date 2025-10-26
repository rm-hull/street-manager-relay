package routes

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/rm-hull/street-manager-relay/generated"
	"github.com/rm-hull/street-manager-relay/internal"
	"github.com/rm-hull/street-manager-relay/models"
)

func HandleSNSMessage(repo *internal.DbRepository, certManager internal.CertManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		messageType := c.GetHeader("x-amz-sns-message-type")
		if messageType == "" {
			_ = c.Error(errors.New("missing x-amz-sns-message-type header"))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing x-amz-sns-message-type header"})
			return
		}

		bodyBytes, err := c.GetRawData()
		if err != nil {
			_ = c.Error(errors.Wrap(err, "error reading request body"))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}

		var body internal.SNSMessage
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			_ = c.Error(errors.Wrap(err, "error parsing JSON"))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		valid, err := internal.IsValidSignature(&body, certManager)
		if err != nil {
			_ = c.Error(errors.Wrap(err, "signature validation failed"))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Signature validation failed"})
			return
		}

		if !valid {
			_ = c.Error(errors.New("message signature is not valid"))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Message signature is not valid"})
			return
		}

		if err := handleMessage(repo, &body); err != nil {
			_ = c.Error(errors.Wrap(err, "failed to handle message "))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to handle message"})
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
		return errors.Wrap(err, "failed to confirm subscription")
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return errors.Newf("subscription confirmation failed with HTTP %d", resp.StatusCode)
	}

	log.Println("Subscription confirmed")
	return nil
}

func handleNotification(repo *internal.DbRepository, body *internal.SNSMessage) error {
	event, err := generated.UnmarshalEventNotifierMessage([]byte(body.Message))
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal event")
	}

	batch, err := repo.BatchUpsert()
	if err != nil {
		return errors.Wrap(err, "failed to create batch upserter")
	}

	if _, err = batch.Upsert(models.NewEventFrom(event)); err != nil {
		return errors.Wrap(batch.Abort(err), "failed to upsert")
	}

	return batch.Done()
}
