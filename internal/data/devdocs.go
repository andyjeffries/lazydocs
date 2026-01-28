package data

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/lazydocs/lazydocs/internal/model"
)

const (
	// DevDocs API endpoints
	manifestURL = "https://devdocs.io/docs.json"
	docsBaseURL = "https://documents.devdocs.io"
)

// Client handles DevDocs API requests
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new DevDocs client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchManifest downloads the list of available docsets
func (c *Client) FetchManifest() (model.Manifest, error) {
	resp, err := c.httpClient.Get(manifestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manifest request failed with status %d", resp.StatusCode)
	}

	var manifest model.Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("failed to decode manifest: %w", err)
	}

	return manifest, nil
}

// DocsetData represents the raw docset data from DevDocs
type DocsetData struct {
	Entries []DocsetEntry `json:"entries"`
}

// DocsetEntry represents a single entry in the docset
type DocsetEntry struct {
	Name string `json:"name"` // Display name
	Path string `json:"path"` // Path identifier
	Type string `json:"type"` // Category/type
}

// FetchDocset downloads the documentation database for a docset
func (c *Client) FetchDocset(slug string, progress func(downloaded, total int64)) ([]byte, error) {
	url := fmt.Sprintf("%s/%s/db.json", docsBaseURL, slug)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch docset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("docset request failed with status %d", resp.StatusCode)
	}

	total := resp.ContentLength

	// Read with progress tracking
	var data []byte
	buf := make([]byte, 32*1024) // 32KB buffer
	var downloaded int64

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			data = append(data, buf[:n]...)
			downloaded += int64(n)
			if progress != nil {
				progress(downloaded, total)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read docset data: %w", err)
		}
	}

	return data, nil
}

// FetchIndex fetches the index.json for a docset (contains entry list)
func (c *Client) FetchIndex(slug string) (*DocsetData, error) {
	url := fmt.Sprintf("%s/%s/index.json", docsBaseURL, slug)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("index request failed with status %d", resp.StatusCode)
	}

	var data DocsetData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode index: %w", err)
	}

	return &data, nil
}
