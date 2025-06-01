package client

import (
	"context"
	"fmt"
)

// Ping checks the API connection and initializes telemetry
func (c *MemoryClient) Ping(ctx context.Context) error {
	response, err := c.fetchWithErrorHandling(ctx, "GET", "/v1/ping/", nil)
	if err != nil {
		return fmt.Errorf("failed to ping server: %w", err)
	}

	responseMap, ok := response.(map[string]interface{})
	if !ok {
		return NewAPIError("Invalid response format from ping endpoint", 0, "")
	}

	status, ok := responseMap["status"].(string)
	if !ok || status != "ok" {
		message, _ := responseMap["message"].(string)
		if message == "" {
			message = "API Key is invalid"
		}
		return NewAPIError(message, 0, "")
	}

	// Update client configuration from response
	if orgID, exists := responseMap["org_id"]; exists && c.organizationID == nil {
		c.organizationID = orgID
	}
	if projectID, exists := responseMap["project_id"]; exists && c.projectID == nil {
		c.projectID = projectID
	}
	if userEmail, exists := responseMap["user_email"].(string); exists {
		c.telemetryID = userEmail
	}

	return nil
}

// Add creates new memories from messages
func (c *MemoryClient) Add(ctx context.Context, messages []Message, options ...MemoryOptions) ([]Memory, error) {
	if c.telemetryID == "" {
		if err := c.Ping(ctx); err != nil {
			return nil, err
		}
	}

	c.validateOrgProject()

	// Use first options or empty options
	opts := MemoryOptions{}
	if len(options) > 0 {
		opts = options[0]
	}

	// Set organization/project info
	if c.organizationName != nil && c.projectName != nil {
		opts.OrgName = c.organizationName
		opts.ProjectName = c.projectName
	}

	if c.organizationID != nil && c.projectID != nil {
		opts.OrgID = c.organizationID
		opts.ProjectID = c.projectID
		// Remove deprecated fields if using new ones
		opts.OrgName = nil
		opts.ProjectName = nil
	}

	// Handle API version
	if opts.APIVersion != nil {
		version := string(*opts.APIVersion)
		opts.Version = (*APIVersion)(&version)
	}

	payload := c.preparePayload(messages, opts)

	response, err := c.fetchWithErrorHandling(ctx, "POST", "/v1/memories/", payload)
	if err != nil {
		return nil, err
	}

	// Parse response to []Memory
	var memories []Memory
	if err := parseResponse(response, &memories); err != nil {
		return nil, err
	}

	return memories, nil
}

// Update modifies an existing memory
func (c *MemoryClient) Update(ctx context.Context, memoryID, message string) ([]Memory, error) {
	if c.telemetryID == "" {
		if err := c.Ping(ctx); err != nil {
			return nil, err
		}
	}

	c.validateOrgProject()

	payload := map[string]interface{}{
		"text": message,
	}

	endpoint := fmt.Sprintf("/v1/memories/%s/", memoryID)
	response, err := c.fetchWithErrorHandling(ctx, "PUT", endpoint, payload)
	if err != nil {
		return nil, err
	}

	var memories []Memory
	if err := parseResponse(response, &memories); err != nil {
		return nil, err
	}

	return memories, nil
}

// Get retrieves a specific memory by ID
func (c *MemoryClient) Get(ctx context.Context, memoryID string) (*Memory, error) {
	if c.telemetryID == "" {
		if err := c.Ping(ctx); err != nil {
			return nil, err
		}
	}

	endpoint := fmt.Sprintf("/v1/memories/%s/", memoryID)
	response, err := c.fetchWithErrorHandling(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var memory Memory
	if err := parseResponse(response, &memory); err != nil {
		return nil, err
	}

	return &memory, nil
}

// GetAll retrieves all memories with optional filters
func (c *MemoryClient) GetAll(ctx context.Context, options ...SearchOptions) ([]Memory, error) {
	if c.telemetryID == "" {
		if err := c.Ping(ctx); err != nil {
			return nil, err
		}
	}

	c.validateOrgProject()

	// Use first options or empty options
	opts := SearchOptions{}
	if len(options) > 0 {
		opts = options[0]
	}

	// Set organization/project info
	if c.organizationName != nil && c.projectName != nil {
		opts.OrgName = c.organizationName
		opts.ProjectName = c.projectName
	}

	if c.organizationID != nil && c.projectID != nil {
		opts.OrgID = c.organizationID
		opts.ProjectID = c.projectID
		opts.OrgName = nil
		opts.ProjectName = nil
	}

	var endpoint string
	var method string
	var requestBody interface{}

	// Handle pagination
	paginationParams := ""
	if opts.Page != nil && opts.PageSize != nil {
		paginationParams = fmt.Sprintf("page=%d&page_size=%d", *opts.Page, *opts.PageSize)
	}

	if opts.APIVersion != nil && *opts.APIVersion == APIVersionV2 {
		// V2 API uses POST
		method = "POST"
		if paginationParams != "" {
			endpoint = fmt.Sprintf("/v2/memories/?%s", paginationParams)
		} else {
			endpoint = "/v2/memories/"
		}
		// Prepare request body for V2
		requestBody = map[string]interface{}{}
		if opts.OrgID != nil {
			requestBody.(map[string]interface{})["org_id"] = opts.OrgID
		}
		if opts.ProjectID != nil {
			requestBody.(map[string]interface{})["project_id"] = opts.ProjectID
		}
	} else {
		// V1 API uses GET with query parameters
		method = "GET"
		params := c.prepareParams(opts.MemoryOptions)
		queryString := params.Encode()
		if paginationParams != "" && queryString != "" {
			endpoint = fmt.Sprintf("/v1/memories/?%s&%s", queryString, paginationParams)
		} else if paginationParams != "" {
			endpoint = fmt.Sprintf("/v1/memories/?%s", paginationParams)
		} else if queryString != "" {
			endpoint = fmt.Sprintf("/v1/memories/?%s", queryString)
		} else {
			endpoint = "/v1/memories/"
		}
	}

	response, err := c.fetchWithErrorHandling(ctx, method, endpoint, requestBody)
	if err != nil {
		return nil, err
	}

	var memories []Memory
	if err := parseResponse(response, &memories); err != nil {
		return nil, err
	}

	return memories, nil
}

// Search searches for memories matching a query
func (c *MemoryClient) Search(ctx context.Context, query string, options ...SearchOptions) ([]Memory, error) {
	if c.telemetryID == "" {
		if err := c.Ping(ctx); err != nil {
			return nil, err
		}
	}

	c.validateOrgProject()

	opts := SearchOptions{}
	if len(options) > 0 {
		opts = options[0]
	}

	payload := map[string]interface{}{
		"query": query,
	}

	// Set organization/project info
	if c.organizationName != nil && c.projectName != nil {
		payload["org_name"] = *c.organizationName
		payload["project_name"] = *c.projectName
	}

	if c.organizationID != nil && c.projectID != nil {
		payload["org_id"] = c.organizationID
		payload["project_id"] = c.projectID
		delete(payload, "org_name")
		delete(payload, "project_name")
	}

	// Add search options to payload
	addSearchOptionsToPayload(payload, opts)

	endpoint := "/v1/memories/search/"
	if opts.APIVersion != nil && *opts.APIVersion == APIVersionV2 {
		endpoint = "/v2/memories/search/"
	}

	response, err := c.fetchWithErrorHandling(ctx, "POST", endpoint, payload)
	if err != nil {
		return nil, err
	}

	var memories []Memory
	if err := parseResponse(response, &memories); err != nil {
		return nil, err
	}

	return memories, nil
}

// Delete removes a specific memory
func (c *MemoryClient) Delete(ctx context.Context, memoryID string) (*MessageResponse, error) {
	if c.telemetryID == "" {
		if err := c.Ping(ctx); err != nil {
			return nil, err
		}
	}

	endpoint := fmt.Sprintf("/v1/memories/%s/", memoryID)
	response, err := c.fetchWithErrorHandling(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result MessageResponse
	if err := parseResponse(response, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteAll removes all memories matching the filter criteria
func (c *MemoryClient) DeleteAll(ctx context.Context, options ...MemoryOptions) (*MessageResponse, error) {
	if c.telemetryID == "" {
		if err := c.Ping(ctx); err != nil {
			return nil, err
		}
	}

	c.validateOrgProject()

	opts := MemoryOptions{}
	if len(options) > 0 {
		opts = options[0]
	}

	// Set organization/project info
	if c.organizationName != nil && c.projectName != nil {
		opts.OrgName = c.organizationName
		opts.ProjectName = c.projectName
	}

	if c.organizationID != nil && c.projectID != nil {
		opts.OrgID = c.organizationID
		opts.ProjectID = c.projectID
		opts.OrgName = nil
		opts.ProjectName = nil
	}

	params := c.prepareParams(opts)
	endpoint := fmt.Sprintf("/v1/memories/?%s", params.Encode())

	response, err := c.fetchWithErrorHandling(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result MessageResponse
	if err := parseResponse(response, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Helper function to add search options to payload
func addSearchOptionsToPayload(payload map[string]interface{}, opts SearchOptions) {
	// Add MemoryOptions fields first
	if opts.UserID != nil {
		payload["user_id"] = *opts.UserID
	}
	if opts.AgentID != nil {
		payload["agent_id"] = *opts.AgentID
	}
	if opts.AppID != nil {
		payload["app_id"] = *opts.AppID
	}
	if opts.RunID != nil {
		payload["run_id"] = *opts.RunID
	}
	if opts.Metadata != nil {
		payload["metadata"] = opts.Metadata
	}
	if opts.Filters != nil {
		payload["filters"] = opts.Filters
	}

	// Add search-specific options
	if opts.Limit != nil {
		payload["limit"] = *opts.Limit
	}
	if opts.EnableGraph != nil {
		payload["enable_graph"] = *opts.EnableGraph
	}
	if opts.Threshold != nil {
		payload["threshold"] = *opts.Threshold
	}
	if opts.TopK != nil {
		payload["top_k"] = *opts.TopK
	}
	if opts.OnlyMetadataBasedSearch != nil {
		payload["only_metadata_based_search"] = *opts.OnlyMetadataBasedSearch
	}
	if opts.KeywordSearch != nil {
		payload["keyword_search"] = *opts.KeywordSearch
	}
	if opts.Fields != nil {
		payload["fields"] = opts.Fields
	}
	if opts.Categories != nil {
		payload["categories"] = opts.Categories
	}
	if opts.Rerank != nil {
		payload["rerank"] = *opts.Rerank
	}
}

// BatchUpdate updates multiple memories in a single request
func (c *MemoryClient) BatchUpdate(ctx context.Context, memories []MemoryUpdateBody) (string, error) {
	if c.telemetryID == "" {
		if err := c.Ping(ctx); err != nil {
			return "", err
		}
	}

	memoriesBody := make([]map[string]interface{}, len(memories))
	for i, memory := range memories {
		memoriesBody[i] = map[string]interface{}{
			"memory_id": memory.MemoryID,
			"text":      memory.Text,
		}
	}

	payload := map[string]interface{}{
		"memories": memoriesBody,
	}

	response, err := c.fetchWithErrorHandling(ctx, "PUT", "/v1/batch/", payload)
	if err != nil {
		return "", err
	}

	// The response is expected to be a string
	if result, ok := response.(string); ok {
		return result, nil
	}

	// If not a string, try to parse as a map and get message
	if responseMap, ok := response.(map[string]interface{}); ok {
		if message, exists := responseMap["message"].(string); exists {
			return message, nil
		}
	}

	return "Batch update completed", nil
}

// BatchDelete deletes multiple memories in a single request
func (c *MemoryClient) BatchDelete(ctx context.Context, memoryIDs []string) (string, error) {
	if c.telemetryID == "" {
		if err := c.Ping(ctx); err != nil {
			return "", err
		}
	}

	memoriesBody := make([]map[string]interface{}, len(memoryIDs))
	for i, memoryID := range memoryIDs {
		memoriesBody[i] = map[string]interface{}{
			"memory_id": memoryID,
		}
	}

	payload := map[string]interface{}{
		"memories": memoriesBody,
	}

	response, err := c.fetchWithErrorHandling(ctx, "DELETE", "/v1/batch/", payload)
	if err != nil {
		return "", err
	}

	// The response is expected to be a string
	if result, ok := response.(string); ok {
		return result, nil
	}

	// If not a string, try to parse as a map and get message
	if responseMap, ok := response.(map[string]interface{}); ok {
		if message, exists := responseMap["message"].(string); exists {
			return message, nil
		}
	}

	return "Batch delete completed", nil
}

// History retrieves the change history for a specific memory
func (c *MemoryClient) History(ctx context.Context, memoryID string) ([]MemoryHistory, error) {
	if c.telemetryID == "" {
		if err := c.Ping(ctx); err != nil {
			return nil, err
		}
	}

	endpoint := fmt.Sprintf("/v1/memories/%s/history/", memoryID)
	response, err := c.fetchWithErrorHandling(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var history []MemoryHistory
	if err := parseResponse(response, &history); err != nil {
		return nil, err
	}

	return history, nil
}

// Users retrieves all users/entities
func (c *MemoryClient) Users(ctx context.Context) (*AllUsers, error) {
	if c.telemetryID == "" {
		if err := c.Ping(ctx); err != nil {
			return nil, err
		}
	}

	c.validateOrgProject()

	options := MemoryOptions{}
	if c.organizationName != nil && c.projectName != nil {
		options.OrgName = c.organizationName
		options.ProjectName = c.projectName
	}

	if c.organizationID != nil && c.projectID != nil {
		options.OrgID = c.organizationID
		options.ProjectID = c.projectID
		options.OrgName = nil
		options.ProjectName = nil
	}

	params := c.prepareParams(options)
	endpoint := fmt.Sprintf("/v1/entities/?%s", params.Encode())

	response, err := c.fetchWithErrorHandling(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var users AllUsers
	if err := parseResponse(response, &users); err != nil {
		return nil, err
	}

	return &users, nil
}

// DeleteUser deletes a user entity (deprecated - use DeleteUsers instead)
func (c *MemoryClient) DeleteUser(ctx context.Context, data DeleteUserData) (*MessageResponse, error) {
	if c.telemetryID == "" {
		if err := c.Ping(ctx); err != nil {
			return nil, err
		}
	}

	entityType := data.EntityType
	if entityType == "" {
		entityType = "user"
	}

	endpoint := fmt.Sprintf("/v1/entities/%s/%d/", entityType, data.EntityID)
	response, err := c.fetchWithErrorHandling(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result MessageResponse
	if err := parseResponse(response, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteUsers deletes users based on the provided parameters
func (c *MemoryClient) DeleteUsers(ctx context.Context, params ...DeleteUsersParams) (*MessageResponse, error) {
	if c.telemetryID == "" {
		if err := c.Ping(ctx); err != nil {
			return nil, err
		}
	}

	c.validateOrgProject()

	var toDelete []map[string]string

	// Use first params or empty params
	deleteParams := DeleteUsersParams{}
	if len(params) > 0 {
		deleteParams = params[0]
	}

	// Determine what to delete based on parameters
	if deleteParams.UserID != nil {
		toDelete = []map[string]string{{"type": "user", "name": *deleteParams.UserID}}
	} else if deleteParams.AgentID != nil {
		toDelete = []map[string]string{{"type": "agent", "name": *deleteParams.AgentID}}
	} else if deleteParams.AppID != nil {
		toDelete = []map[string]string{{"type": "app", "name": *deleteParams.AppID}}
	} else if deleteParams.RunID != nil {
		toDelete = []map[string]string{{"type": "run", "name": *deleteParams.RunID}}
	} else {
		// Delete all entities
		entities, err := c.Users(ctx)
		if err != nil {
			return nil, err
		}

		toDelete = make([]map[string]string, len(entities.Results))
		for i, entity := range entities.Results {
			toDelete[i] = map[string]string{
				"type": entity.Type,
				"name": entity.Name,
			}
		}
	}

	if len(toDelete) == 0 {
		return nil, fmt.Errorf("no entities to delete")
	}

	requestOptions := MemoryOptions{}
	if c.organizationName != nil && c.projectName != nil {
		requestOptions.OrgName = c.organizationName
		requestOptions.ProjectName = c.projectName
	}

	if c.organizationID != nil && c.projectID != nil {
		requestOptions.OrgID = c.organizationID
		requestOptions.ProjectID = c.projectID
		requestOptions.OrgName = nil
		requestOptions.ProjectName = nil
	}

	// Delete each entity
	for _, entity := range toDelete {
		endpoint := fmt.Sprintf("/v2/entities/%s/%s/", entity["type"], entity["name"])
		params := c.prepareParams(requestOptions)
		if params.Encode() != "" {
			endpoint += "?" + params.Encode()
		}

		_, err := c.fetchWithErrorHandling(ctx, "DELETE", endpoint, nil)
		if err != nil {
			return nil, NewAPIError(
				fmt.Sprintf("Failed to delete %s %s: %s", entity["type"], entity["name"], err.Error()),
				0, "",
			)
		}
	}

	message := "All users, agents, apps and runs deleted."
	if deleteParams.UserID != nil || deleteParams.AgentID != nil || deleteParams.AppID != nil || deleteParams.RunID != nil {
		message = "Entity deleted successfully."
	}

	return &MessageResponse{Message: message}, nil
}
