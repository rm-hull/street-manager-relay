package internal

import (
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/cockroachdb/errors"
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
	certificate, err, _ := memoize.Call(cm.cache, certURL, func() (string, error) {
		return cm.download(certURL)
	})
	return certificate, errors.Wrapf(err, "unable to download from: %s", certURL)
}

func (cm *CachedCertManager) verifyMessageSignatureURL(certURL string) error {
	parsedURL, err := url.Parse(certURL)
	if err != nil {
		return errors.Wrapf(err, "invalid URL: %s", certURL)
	}

	if parsedURL.Scheme != "https" {
		return errors.New("SigningCertURL was not using HTTPS")
	}
	return nil
}

func (cm *CachedCertManager) download(certURL string) (string, error) {
	log.Printf("Downloading certificate: %s", certURL)
	if err := cm.verifyMessageSignatureURL(certURL); err != nil {
		return "", errors.Wrap(err, "failed to verify signature URL")
	}

	resp, err := http.Get(certURL)
	if err != nil {
		return "", errors.Wrap(err, "error fetching certificate")
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", errors.Newf("error fetching certificate: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "error reading certificate response: %w")
	}

	return string(body), nil
}
