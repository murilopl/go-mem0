package client

import (
	"testing"
)

func TestValidateAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		wantErr  bool
		errField string
	}{
		{
			name:     "valid API key",
			apiKey:   "valid-api-key",
			wantErr:  false,
		},
		{
			name:     "empty API key",
			apiKey:   "",
			wantErr:  true,
			errField: "apiKey",
		},
		{
			name:     "whitespace only API key",
			apiKey:   "   ",
			wantErr:  true,
			errField: "apiKey",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAPIKey(tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAPIKey() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				if validationErr, ok := err.(*ValidationError); ok {
					if validationErr.Field != tt.errField {
						t.Errorf("validateAPIKey() error field = %v, want %v", validationErr.Field, tt.errField)
					}
				}
			}
		})
	}
}

func TestNewMemoryClient(t *testing.T) {
	tests := []struct {
		name    string
		options ClientOptions
		wantErr bool
	}{
		{
			name: "valid options",
			options: ClientOptions{
				APIKey: "test-api-key",
			},
			wantErr: false,
		},
		{
			name: "valid options with host",
			options: ClientOptions{
				APIKey: "test-api-key",
				Host:   stringPtr("https://custom.api.com"),
			},
			wantErr: false,
		},
		{
			name: "invalid API key",
			options: ClientOptions{
				APIKey: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewMemoryClient(tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMemoryClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewMemoryClient() returned nil client")
			}
			if !tt.wantErr {
				// Validate client configuration
				if client.apiKey != tt.options.APIKey {
					t.Errorf("NewMemoryClient() apiKey = %v, want %v", client.apiKey, tt.options.APIKey)
				}
				expectedHost := "https://api.mem0.ai"
				if tt.options.Host != nil {
					expectedHost = *tt.options.Host
				}
				if client.host != expectedHost {
					t.Errorf("NewMemoryClient() host = %v, want %v", client.host, expectedHost)
				}
			}
		})
	}
}

func TestPreparePayload(t *testing.T) {
	client := &MemoryClient{}
	
	messages := []Message{
		{Role: "user", Content: "test message"},
	}
	
	userID := "test-user"
	apiVersion := APIVersionV1
	
	options := MemoryOptions{
		UserID:     &userID,
		APIVersion: &apiVersion,
	}
	
	payload := client.preparePayload(messages, options)
	
	// Validate messages are included
	if messagesValue, ok := payload["messages"]; !ok {
		t.Error("Payload should include messages")
	} else if len(messagesValue.([]Message)) != 1 {
		t.Error("Payload should include all messages")
	}
	
	// Validate user_id is included
	if userIDValue, ok := payload["user_id"]; !ok {
		t.Error("Payload should include user_id")
	} else if userIDValue != userID {
		t.Errorf("Payload user_id = %v, want %v", userIDValue, userID)
	}
	
	// Validate api_version is included
	if apiVersionValue, ok := payload["api_version"]; !ok {
		t.Error("Payload should include api_version")
	} else if apiVersionValue != apiVersion {
		t.Errorf("Payload api_version = %v, want %v", apiVersionValue, apiVersion)
	}
}

func TestPrepareParams(t *testing.T) {
	client := &MemoryClient{}
	
	userID := "test-user"
	orgID := "test-org"
	
	options := MemoryOptions{
		UserID: &userID,
		OrgID:  orgID,
	}
	
	params := client.prepareParams(options)
	
	// Validate user_id parameter
	if userIDValues := params["user_id"]; len(userIDValues) != 1 || userIDValues[0] != userID {
		t.Errorf("Params user_id = %v, want %v", userIDValues, []string{userID})
	}
	
	// Validate org_id parameter
	if orgIDValues := params["org_id"]; len(orgIDValues) != 1 || orgIDValues[0] != "test-org" {
		t.Errorf("Params org_id = %v, want %v", orgIDValues, []string{"test-org"})
	}
}

func TestErrorTypes(t *testing.T) {
	t.Run("APIError", func(t *testing.T) {
		err := NewAPIError("test message", 400, "response body")
		
		if err.Message != "test message" {
			t.Errorf("APIError Message = %v, want %v", err.Message, "test message")
		}
		if err.StatusCode != 400 {
			t.Errorf("APIError StatusCode = %v, want %v", err.StatusCode, 400)
		}
		if err.Body != "response body" {
			t.Errorf("APIError Body = %v, want %v", err.Body, "response body")
		}
		
		expectedError := "API request failed (status 400): test message"
		if err.Error() != expectedError {
			t.Errorf("APIError Error() = %v, want %v", err.Error(), expectedError)
		}
	})
	
	t.Run("ValidationError", func(t *testing.T) {
		err := NewValidationError("apiKey", "is required")
		
		if err.Field != "apiKey" {
			t.Errorf("ValidationError Field = %v, want %v", err.Field, "apiKey")
		}
		if err.Message != "is required" {
			t.Errorf("ValidationError Message = %v, want %v", err.Message, "is required")
		}
		
		expectedError := "validation error for field 'apiKey': is required"
		if err.Error() != expectedError {
			t.Errorf("ValidationError Error() = %v, want %v", err.Error(), expectedError)
		}
	})
}

func TestRandomString(t *testing.T) {
	// Test that randomString generates strings
	str1 := randomString()
	str2 := randomString()
	
	if str1 == "" {
		t.Error("randomString() should not return empty string")
	}
	if str2 == "" {
		t.Error("randomString() should not return empty string")
	}
	if str1 == str2 {
		t.Error("randomString() should generate different strings")
	}
	if len(str1) != 20 {
		t.Errorf("randomString() length = %v, want %v", len(str1), 20)
	}
}

func TestValidateMemoryObject(t *testing.T) {
	memory := Memory{
		ID:     "test-id",
		Memory: stringPtr("test memory content"),
		UserID: stringPtr("test-user"),
	}
	
	// This test validates that our validation helper doesn't panic
	// and properly identifies valid memory objects
	validateMemoryObject(t, memory, "test-user")
	
	// Test with invalid user ID should be handled by the test framework
	// We can't easily test t.Error calls without more complex setup
}