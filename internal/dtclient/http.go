package dtclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ShinyaT1994/release-catalog/internal/shared/apperror"
)

// HTTPClient implements Client using real DT REST API
type HTTPClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewHTTPClient(baseURL, apiKey string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *HTTPClient) GetProject(ctx context.Context, uuid string) (*Project, error) {
	resp, err := c.doGet(ctx, fmt.Sprintf("/api/v1/project/%s", uuid))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, apperror.New(apperror.CodeRootProjectNotFound, fmt.Sprintf("project %s not found", uuid))
	}
	if resp.StatusCode != http.StatusOK {
		return nil, apperror.New(apperror.CodeDTUnavailable, fmt.Sprintf("DT returned status %d", resp.StatusCode))
	}

	var p Project
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *HTTPClient) ProjectExists(ctx context.Context, uuid string) (bool, error) {
	_, err := c.GetProject(ctx, uuid)
	if err != nil {
		if appErr, ok := err.(*apperror.Error); ok && appErr.Code == apperror.CodeRootProjectNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *HTTPClient) GetBOM(ctx context.Context, projectUUID string) (*CycloneDXBOM, error) {
	resp, err := c.doGet(ctx, fmt.Sprintf("/api/v1/bom/cyclonedx/project/%s?format=json", projectUUID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, apperror.New(apperror.CodeRootProjectNotFound, fmt.Sprintf("BOM not found for project %s", projectUUID))
	}
	if resp.StatusCode != http.StatusOK {
		return nil, apperror.New(apperror.CodeDTUnavailable, fmt.Sprintf("DT returned status %d", resp.StatusCode))
	}

	var bom CycloneDXBOM
	if err := json.NewDecoder(resp.Body).Decode(&bom); err != nil {
		return nil, err
	}
	return &bom, nil
}

func (c *HTTPClient) GetVulnerabilities(ctx context.Context, projectUUID string) ([]*Vulnerability, error) {
	resp, err := c.doGet(ctx, fmt.Sprintf("/api/v1/vulnerability/project/%s", projectUUID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apperror.New(apperror.CodeDTUnavailable, fmt.Sprintf("DT returned status %d", resp.StatusCode))
	}

	var vulns []*Vulnerability
	if err := json.NewDecoder(resp.Body).Decode(&vulns); err != nil {
		return nil, err
	}
	return vulns, nil
}

func (c *HTTPClient) doGet(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-Api-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, apperror.New(apperror.CodeDTUnavailable, fmt.Sprintf("failed to connect to Dependency-Track: %v", err))
	}
	return resp, nil
}
