package internal

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// SNSMessage represents the structure of an SNS message
type SNSMessage struct {
	Type             string `json:"Type"`
	MessageId        string `json:"MessageId"`
	TopicArn         string `json:"TopicArn"`
	Subject          string `json:"Subject,omitempty"`
	Message          string `json:"Message"`
	Timestamp        string `json:"Timestamp"`
	SignatureVersion string `json:"SignatureVersion"`
	Signature        string `json:"Signature"`
	SigningCertURL   string `json:"SigningCertURL"`
	SubscribeURL     string `json:"SubscribeURL,omitempty"`
	Token            string `json:"Token,omitempty"`
}

func UnmarshalSNSMessage(data []byte) (SNSMessage, error) {
	var r SNSMessage
	err := json.Unmarshal(data, &r)
	return r, err
}

func IsValidSignature(body *SNSMessage, certManager CertManager) (bool, error) {
	if err := verifyMessageSignatureVersion(body.SignatureVersion); err != nil {
		return false, err
	}

	certificate, err := certManager.Download(body.SigningCertURL)
	if err != nil {
		return false, err
	}

	return validateSignature(body, certificate)
}

func verifyMessageSignatureVersion(version string) error {
	if version != "1" {
		return errors.New("signature verification failed")
	}
	return nil
}

func validateSignature(message *SNSMessage, certificate string) (bool, error) {
	block, _ := pem.Decode([]byte(certificate))
	if block == nil {
		return false, errors.New("failed to parse PEM certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse certificate: %w", err)
	}

	rsaPubKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return false, errors.New("certificate does not contain RSA public key")
	}

	messageToSign := getMessageToSign(message)
	if messageToSign == "" {
		return false, errors.New("unable to build message to sign")
	}

	hash := sha1.Sum([]byte(messageToSign))

	signature, err := base64.StdEncoding.DecodeString(message.Signature)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %w", err)
	}

	err = rsa.VerifyPKCS1v15(rsaPubKey, crypto.SHA1, hash[:], signature)
	return err == nil, nil
}

func getMessageToSign(body *SNSMessage) string {
	switch body.Type {
	case "SubscriptionConfirmation":
		return buildSubscriptionStringToSign(body)
	case "Notification":
		return buildNotificationStringToSign(body)
	default:
		return ""
	}
}

func buildNotificationStringToSign(body *SNSMessage) string {
	var builder strings.Builder

	builder.WriteString("Message\n")
	builder.WriteString(body.Message + "\n")
	builder.WriteString("MessageId\n")
	builder.WriteString(body.MessageId + "\n")

	if body.Subject != "" {
		builder.WriteString("Subject\n")
		builder.WriteString(body.Subject + "\n")
	}

	builder.WriteString("Timestamp\n")
	builder.WriteString(body.Timestamp + "\n")
	builder.WriteString("TopicArn\n")
	builder.WriteString(body.TopicArn + "\n")
	builder.WriteString("Type\n")
	builder.WriteString(body.Type + "\n")

	return builder.String()
}

func buildSubscriptionStringToSign(body *SNSMessage) string {
	var builder strings.Builder

	builder.WriteString("Message\n")
	builder.WriteString(body.Message + "\n")
	builder.WriteString("MessageId\n")
	builder.WriteString(body.MessageId + "\n")
	builder.WriteString("SubscribeURL\n")
	builder.WriteString(body.SubscribeURL + "\n")
	builder.WriteString("Timestamp\n")
	builder.WriteString(body.Timestamp + "\n")
	builder.WriteString("Token\n")
	builder.WriteString(body.Token + "\n")
	builder.WriteString("TopicArn\n")
	builder.WriteString(body.TopicArn + "\n")
	builder.WriteString("Type\n")
	builder.WriteString(body.Type + "\n")

	return builder.String()
}
