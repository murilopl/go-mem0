package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ClientOptions represents configuration options for the MemoryClient
type ClientOptions struct {
	APIKey           string      `json:"apiKey"`
	Host             *string     `json:"host,omitempty"`
	OrganizationName *string     `json:"organizationName,omitempty"` // Deprecated
	ProjectName      *string     `json:"projectName,omitempty"`      // Deprecated
	OrganizationID   interface{} `json:"organizationId,omitempty"`   // string or number
	ProjectID        interface{} `json:"projectId,omitempty"`        // string or number
}

// MemoryClient represents the main client for interacting with the Mem0 API
type MemoryClient struct {
	apiKey           string
	host             string
	organizationName *string
	projectName      *string
	organizationID   interface{}
	projectID        interface{}
	headers          map[string]string
	httpClient       *http.Client
	telemetryID      string
}

// NewMemoryClient creates a new MemoryClient instance
func NewMemoryClient(options ClientOptions) (*MemoryClient, error) {
	if err := validateAPIKey(options.APIKey); err != nil {
		return nil, err
	}

	host := "https://api.mem0.ai"
	if options.Host != nil {
		host = *options.Host
	}

	client := &MemoryClient{
		apiKey:           options.APIKey,
		host:             host,
		organizationName: options.OrganizationName,
		projectName:      options.ProjectName,
		organizationID:   options.OrganizationID,
		projectID:        options.ProjectID,
		headers: map[string]string{
			"Authorization": fmt.Sprintf("Token %s", options.APIKey),
			"Content-Type":  "application/json",
		},
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		telemetryID: "",
	}

	// Initialize the client
	if err := client.initializeClient(context.Background()); err != nil {
		// Log error but don't fail initialization
		fmt.Printf("Failed to initialize client: %v\n", err)
	}

	return client, nil
}

// validateAPIKey validates the API key
func validateAPIKey(apiKey string) error {
	if apiKey == "" {
		return NewValidationError("apiKey", "Mem0 API key is required")
	}
	if strings.TrimSpace(apiKey) == "" {
		return NewValidationError("apiKey", "Mem0 API key cannot be empty")
	}
	return nil
}

// validateOrgProject validates organization and project configuration
func (c *MemoryClient) validateOrgProject() {
	// Check for organizationName/projectName pair
	if (c.organizationName == nil && c.projectName != nil) ||
		(c.organizationName != nil && c.projectName == nil) {
		fmt.Println("Warning: Both organizationName and projectName must be provided together when using either. This will be removed from version 1.0.40. Note that organizationName/projectName are being deprecated in favor of organizationId/projectId.")
	}

	// Check for organizationId/projectId pair
	if (c.organizationID == nil && c.projectID != nil) ||
		(c.organizationID != nil && c.projectID == nil) {
		fmt.Println("Warning: Both organizationId and projectId must be provided together when using either. This will be removed from version 1.0.40.")
	}
}

// initializeClient initializes the client by pinging the server
func (c *MemoryClient) initializeClient(ctx context.Context) error {
	// Generate telemetry ID
	if err := c.Ping(ctx); err != nil {
		return err
	}

	c.validateOrgProject()
	return nil
}

// fetchWithErrorHandling makes HTTP requests with error handling
func (c *MemoryClient) fetchWithErrorHandling(ctx context.Context, method, endpoint string, body interface{}) (interface{}, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	url := fmt.Sprintf("%s%s", c.host, endpoint)
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}
	if c.telemetryID != "" {
		req.Header.Set("Mem0-User-ID", c.telemetryID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, NewAPIError(string(respBody), resp.StatusCode, string(respBody))
	}

	var result interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	return result, nil
}

// preparePayload combines messages with options for API requests
func (c *MemoryClient) preparePayload(messages []Message, options MemoryOptions) map[string]interface{} {
	payload := make(map[string]interface{})
	payload["messages"] = messages

	// Add all non-nil options to payload
	if options.APIVersion != nil {
		payload["api_version"] = *options.APIVersion
	}
	if options.Version != nil {
		payload["version"] = *options.Version
	}
	if options.UserID != nil {
		payload["user_id"] = *options.UserID
	}
	if options.AgentID != nil {
		payload["agent_id"] = *options.AgentID
	}
	if options.AppID != nil {
		payload["app_id"] = *options.AppID
	}
	if options.RunID != nil {
		payload["run_id"] = *options.RunID
	}
	if options.Metadata != nil {
		payload["metadata"] = options.Metadata
	}
	if options.Filters != nil {
		payload["filters"] = options.Filters
	}
	if options.OrgName != nil {
		payload["org_name"] = *options.OrgName
	}
	if options.ProjectName != nil {
		payload["project_name"] = *options.ProjectName
	}
	if options.OrgID != nil {
		payload["org_id"] = options.OrgID
	}
	if options.ProjectID != nil {
		payload["project_id"] = options.ProjectID
	}
	if options.Infer != nil {
		payload["infer"] = *options.Infer
	}
	if options.EnableGraph != nil {
		payload["enable_graph"] = *options.EnableGraph
	}
	if options.CustomCategories != nil {
		payload["custom_categories"] = options.CustomCategories
	}
	if options.CustomInstructions != nil {
		payload["custom_instructions"] = *options.CustomInstructions
	}
	if options.OutputFormat != nil {
		payload["output_format"] = *options.OutputFormat
	}
	if options.AsyncMode != nil {
		payload["async_mode"] = *options.AsyncMode
	}

	return payload
}

// prepareParams converts options to URL parameters
func (c *MemoryClient) prepareParams(options interface{}) url.Values {
	params := url.Values{}

	// Use reflection or type assertion to handle different option types
	switch opts := options.(type) {
	case MemoryOptions:
		if opts.UserID != nil {
			params.Add("user_id", *opts.UserID)
		}
		if opts.AgentID != nil {
			params.Add("agent_id", *opts.AgentID)
		}
		if opts.AppID != nil {
			params.Add("app_id", *opts.AppID)
		}
		if opts.RunID != nil {
			params.Add("run_id", *opts.RunID)
		}
		if opts.OrgName != nil {
			params.Add("org_name", *opts.OrgName)
		}
		if opts.ProjectName != nil {
			params.Add("project_name", *opts.ProjectName)
		}
		if opts.OrgID != nil {
			params.Add("org_id", fmt.Sprintf("%v", opts.OrgID))
		}
		if opts.ProjectID != nil {
			params.Add("project_id", fmt.Sprintf("%v", opts.ProjectID))
		}
	case SearchOptions:
		// Handle SearchOptions by first handling the embedded MemoryOptions
		memOpts := opts.MemoryOptions
		if memOpts.UserID != nil {
			params.Add("user_id", *memOpts.UserID)
		}
		// Add other search-specific parameters as needed
	}

	return params
}
