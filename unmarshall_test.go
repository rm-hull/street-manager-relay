package main

import (
	"os"
	"testing"

	"github.com/rm-hull/street-manager-relay/generated"
	"github.com/rm-hull/street-manager-relay/internal"
)

func TestXxx(t *testing.T) {

	// data, err := os.ReadFile("doc/sample_messages/activity_notification_message.json")
	data, err := os.ReadFile("doc/sample_messages/permit_notification_message.json")
	// data, err := os.ReadFile("doc/sample_messages/section58_notification_message.json")
	if err != nil {
		t.Fatalf("Failed to read sample message: %v", err)
	}

	msg, err := generated.UnmarshalEventNotifierMessage(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}
	t.Logf("Unmarshalled message: %+v", msg)

}

func TestYyy(t *testing.T) {
	raw, err := os.ReadFile("doc/sample_events/notification_event.json")
	if err != nil {
		t.Fatalf("Failed to read sample event: %v", err)
	}

	sns, err := internal.UnmarshalSNSMessage(raw)
	if err != nil {
		t.Fatalf("Failed to unmarshal SNS event: %v", err)
	}
	t.Logf("Unmarshalled event: %+v", sns)

	// isValid, err := internal.IsValidSignature(&sns)
	// if err != nil {
	// 	t.Fatalf("Failed to validate signature: %v", err)
	// }

	// if !isValid {
	// 	t.Fatalf("Signature validation failed")
	// }

	msg, err := generated.UnmarshalEventNotifierMessage([]byte(sns.Message))
	if err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}
	t.Logf("Unmarshalled message: %+v", msg)
}
