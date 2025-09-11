package favicon

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type IconInfo struct {
	Href string
	Src  string
	Size int
}

func resolveURL(base, href string) string {
	if strings.HasPrefix(href, "data:") {
		return href
	}

	u, err := url.Parse(href)
	if err != nil {
		return href
	}
	if u.IsAbs() {
		return href
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return href
	}
	return baseURL.ResolveReference(u).String()
}

// extracts the numeric size from a sizes attribute (e.g., "180x180" -> 180)
func parseSize(sizeStr string) int {
	if sizeStr == "" {
		return 0
	}
	dims := strings.Split(sizeStr, "x")
	if len(dims) != 2 {
		return 0
	}
	size, err := strconv.Atoi(dims[0])
	if err != nil {
		return 0
	}
	return size
}

func Extract(url string) (*IconInfo, error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36")

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from %s: %w", url, err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("failed to close request body: %v", err)
		}
	}()

	if res.StatusCode > 299 {
		return nil, fmt.Errorf("http status response from %s: %s", url, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to build html document: %w", err)
	}

	icons := make([]IconInfo, 0, 10)

	sources := map[string]string{
		"link[rel=\"apple-touch-icon\"]":   "href",
		"link[rel=\"icon\"]":               "href",
		"link[rel=\"shortcut-icon\"]":      "href",
		"link[rel=\"shortcut icon\"]":      "href",
		"meta[name=\"twitter:image\"]":     "content",
		"meta[name=\"og:image\"]":          "content",
		"meta[property=\"twitter:image\"]": "content",
		"meta[property=\"og:image\"]":      "content",
	}

	for selector, attr := range sources {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			if href, ok := s.Attr(attr); ok {
				size := parseSize(s.AttrOr("sizes", "0x0"))
				icons = append(icons, IconInfo{Href: resolveURL(url, href), Size: size, Src: selector})
			}
		})
	}

	if len(icons) == 0 {
		return nil, fmt.Errorf("no icons found for: %s", url)
	}

	sort.Slice(icons, func(i, j int) bool {
		return icons[i].Size > icons[j].Size
	})

	return &icons[0], nil
}
