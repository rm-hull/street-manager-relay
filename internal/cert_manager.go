package internal

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/kofalt/go-memoize"
)

type CertManager interface {
	Download(certURL string) (string, error)
}

type CachedCertManager struct {
	cache *memoize.Memoizer
}

func NewCertManager(cache *memoize.Memoizer) CertManager {
	return &CachedCertManager{cache: cache}
}

func (cm *CachedCertManager) Download(certURL string) (string, error) {
	featureCollection, err, _ := memoize.Call(cm.cache, certURL, func() (string, error) {
		return cm.download(certURL)
	})
	return featureCollection, err
}

func (cm *CachedCertManager) verifyMessageSignatureURL(certURL string) error {
	parsedURL, err := url.Parse(certURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "https" {
		return errors.New("SigningCertURL was not using HTTPS")
	}
	return nil
}

func (cm *CachedCertManager) download(certURL string) (string, error) {
	log.Printf("Downloading certificate: %s", certURL)
	if err := cm.verifyMessageSignatureURL(certURL); err != nil {
		return "", err
	}

	resp, err := http.Get(certURL)
	if err != nil {
		return "", fmt.Errorf("error fetching certificate: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error fetching certificate: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading certificate response: %w", err)
	}

	return string(body), nil
}
