package client

import (
	"encoding/json"
	"fmt"
)

// parseResponse converts a generic response interface to a specific type
func parseResponse(response interface{}, target interface{}) error {
	// Convert response to JSON bytes and then unmarshal to target type
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}