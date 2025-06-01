package client

import (
	"context"
	"math/rand"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

var (
	testClient   *MemoryClient
	testUserID   string
	testMemoryID string
)

// randomString generates a random string similar to the TypeScript version
func randomString() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 20) // 10 + 10 characters like in TS
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// setupTestClient initializes the test client (equivalent to beforeAll)
func setupTestClient(t *testing.T) {
	// Load .env file from parent directory
	_ = godotenv.Load("../.env")

	apiKey := os.Getenv("MEM0_API_KEY")
	if apiKey == "" {
		t.Skip("MEM0_API_KEY environment variable not set")
	}

	var err error
	testClient, err = NewMemoryClient(ClientOptions{
		APIKey: apiKey,
		Host:   stringPtr("https://api.mem0.ai"),
	})
	if err != nil {
		t.Fatalf("Failed to create memory client: %v", err)
	}

	testUserID = randomString()
}

// stringPtr is a helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// validateMemoryObject validates a memory object structure
func validateMemoryObject(t *testing.T, memory Memory, expectedUserID string) {
	t.Helper()

	// Should be a string (memory id)
	if memory.ID == "" {
		t.Error("Memory ID should be a non-empty string")
	}

	// Should be a string (the actual memory content)
	if memory.Memory == nil || *memory.Memory == "" {
		t.Error("Memory content should be a non-empty string")
	}

	// Should be a string and equal to the expectedUserID
	if memory.UserID == nil || *memory.UserID != expectedUserID {
		t.Errorf("Memory user_id should be %s, got %v", expectedUserID, memory.UserID)
	}

	// Categories should be an array of strings or nil
	if memory.Categories != nil {
		for _, category := range memory.Categories {
			if category == "" {
				t.Error("Category should be a non-empty string")
			}
		}
	}

	// Should have valid dates
	if memory.CreatedAt != nil && memory.CreatedAt.IsZero() {
		t.Error("CreatedAt should be a valid date")
	}
	if memory.UpdatedAt != nil && memory.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be a valid date")
	}
}

func TestMemoryClientAPI(t *testing.T) {
	setupTestClient(t)

	messages1 := []Message{
		{Role: "user", Content: "Hey, I am Alex. I'm now a vegetarian."},
		{Role: "assistant", Content: "Hello Alex! Glad to hear!"},
	}

	t.Run("should add messages successfully", func(t *testing.T) {
		ctx := context.Background()
		userID := testUserID
		options := MemoryOptions{UserID: &userID}

		res, err := testClient.Add(ctx, messages1, options)
		if err != nil {
			t.Fatalf("Failed to add messages: %v", err)
		}

		// Validate the response contains an iterable list
		if len(res) == 0 {
			t.Error("Response should contain at least one memory")
		}

		// Validate the fields of the first message in the response
		message := res[0]
		if message.ID == "" {
			t.Error("Message ID should be a non-empty string")
		}
		if message.Data == nil || message.Data.Memory == "" {
			t.Error("Message data.memory should be a non-empty string")
		}
		if message.Event == nil {
			t.Error("Message event should not be nil")
		}

		// Store the memory ID for later use
		testMemoryID = message.ID
	})

	t.Run("should retrieve the specific memory by ID", func(t *testing.T) {
		ctx := context.Background()

		memory, err := testClient.Get(ctx, testMemoryID)
		if err != nil {
			t.Fatalf("Failed to get memory: %v", err)
		}

		validateMemoryObject(t, *memory, testUserID)
	})

	t.Run("should retrieve all users successfully", func(t *testing.T) {
		ctx := context.Background()

		allUsers, err := testClient.Users(ctx)
		if err != nil {
			t.Fatalf("Failed to get users: %v", err)
		}

		// Validate the number of users is a number
		if allUsers.Count < 0 {
			t.Error("User count should be a non-negative number")
		}

		if len(allUsers.Results) > 0 {
			// Validate the structure of the first user
			firstUser := allUsers.Results[0]
			if firstUser.ID == "" {
				t.Error("User ID should be a non-empty string")
			}
			if firstUser.Name == "" {
				t.Error("User name should be a non-empty string")
			}
			if firstUser.CreatedAt.IsZero() {
				t.Error("User created_at should be a valid date")
			}
			if firstUser.UpdatedAt.IsZero() {
				t.Error("User updated_at should be a valid date")
			}
			if firstUser.TotalMemories < 0 {
				t.Error("User total_memories should be a non-negative number")
			}
			if firstUser.Type == "" {
				t.Error("User type should be a non-empty string")
			}
		}

		// Find the user with the name matching testUserID
		var entity *User
		for _, user := range allUsers.Results {
			if user.Name == testUserID {
				entity = &user
				break
			}
		}
		if entity == nil {
			t.Error("Should find entity with matching testUserID")
		}
	})

	t.Run("should retrieve all memories for the user", func(t *testing.T) {
		ctx := context.Background()
		userID := testUserID
		options := SearchOptions{
			MemoryOptions: MemoryOptions{UserID: &userID},
		}

		res3, err := testClient.GetAll(ctx, options)
		if err != nil {
			t.Fatalf("Failed to get all memories: %v", err)
		}

		if len(res3) > 0 {
			// Validate the first memory
			memory := res3[0]
			validateMemoryObject(t, memory, testUserID)
		} else {
			// If there are no memories, assert that the list is empty
			if len(res3) != 0 {
				t.Error("Expected empty list when no memories found")
			}
		}
	})

	t.Run("should search and return results based on provided query and filters (API version 2)", func(t *testing.T) {
		ctx := context.Background()
		userID := testUserID
		threshold := 0.1
		apiVersion := APIVersionV2

		options := SearchOptions{
			MemoryOptions: MemoryOptions{
				APIVersion: &apiVersion,
				Filters: map[string]interface{}{
					"OR": []map[string]interface{}{
						{"user_id": userID},
						{"agent_id": "shopping-assistant"},
					},
				},
			},
			Threshold: &threshold,
		}

		searchResultV2, err := testClient.Search(ctx, "What do you know about me?", options)
		if err != nil {
			t.Fatalf("Failed to search memories (v2): %v", err)
		}

		if len(searchResultV2) > 0 {
			// Validate the first search result
			memory := searchResultV2[0]

			// Should be a string (memory id)
			if memory.ID == "" {
				t.Error("Memory ID should be a non-empty string")
			}

			// Should be a string (the actual memory content)
			if memory.Memory == nil || *memory.Memory == "" {
				t.Error("Memory content should be a non-empty string")
			}

			// Validate user_id or agent_id
			if memory.UserID != nil && *memory.UserID != userID {
				t.Errorf("Memory user_id should be %s, got %s", userID, *memory.UserID)
			}
			if memory.AgentID != nil && *memory.AgentID != "shopping-assistant" {
				t.Errorf("Memory agent_id should be shopping-assistant, got %s", *memory.AgentID)
			}

			// Categories should be an array of strings or nil
			if memory.Categories != nil {
				for _, category := range memory.Categories {
					if category == "" {
						t.Error("Category should be a non-empty string")
					}
				}
			}

			// Should have valid dates
			if memory.CreatedAt != nil && memory.CreatedAt.IsZero() {
				t.Error("CreatedAt should be a valid date")
			}
			if memory.UpdatedAt != nil && memory.UpdatedAt.IsZero() {
				t.Error("UpdatedAt should be a valid date")
			}

			// Should be a number (score)
			if memory.Score == nil {
				t.Error("Memory score should not be nil")
			}
		}
	})

	t.Run("should search and return results based on provided query (API version 1)", func(t *testing.T) {
		ctx := context.Background()
		userID := testUserID
		options := SearchOptions{
			MemoryOptions: MemoryOptions{
				UserID: &userID,
			},
		}

		searchResultV1, err := testClient.Search(ctx, "What is my name?", options)
		if err != nil {
			t.Fatalf("Failed to search memories (v1): %v", err)
		}

		if len(searchResultV1) > 0 {
			// Validate the first search result
			memory := searchResultV1[0]

			// Should be a string (memory id)
			if memory.ID == "" {
				t.Error("Memory ID should be a non-empty string")
			}

			// Should be a string (the actual memory content)
			if memory.Memory == nil || *memory.Memory == "" {
				t.Error("Memory content should be a non-empty string")
			}

			// Should be equal to userID
			if memory.UserID == nil || *memory.UserID != userID {
				t.Errorf("Memory user_id should be %s, got %v", userID, memory.UserID)
			}

			// Categories should be an array of strings or nil
			if memory.Categories != nil {
				for _, category := range memory.Categories {
					if category == "" {
						t.Error("Category should be a non-empty string")
					}
				}
			}

			// Should have valid dates
			if memory.CreatedAt != nil && memory.CreatedAt.IsZero() {
				t.Error("CreatedAt should be a valid date")
			}
			if memory.UpdatedAt != nil && memory.UpdatedAt.IsZero() {
				t.Error("UpdatedAt should be a valid date")
			}

			// Should be a number (score)
			if memory.Score == nil {
				t.Error("Memory score should not be nil")
			}
		}
	})

	t.Run("should retrieve history of a specific memory and validate the fields", func(t *testing.T) {
		ctx := context.Background()

		res22, err := testClient.History(ctx, testMemoryID)
		if err != nil {
			t.Fatalf("Failed to get memory history: %v", err)
		}

		if len(res22) > 0 {
			// Validate the first history entry
			historyEntry := res22[0]

			// Should be a string (history entry id)
			if historyEntry.ID == "" {
				t.Error("History entry ID should be a non-empty string")
			}

			// Should be a string (memory id related to the history entry)
			if historyEntry.MemoryID == "" {
				t.Error("History entry memory_id should be a non-empty string")
			}

			// Should be equal to userID
			if historyEntry.UserID != testUserID {
				t.Errorf("History entry user_id should be %s, got %s", testUserID, historyEntry.UserID)
			}

			// Categories should be an array of strings or nil
			if historyEntry.Categories != nil {
				for _, category := range historyEntry.Categories {
					if category == "" {
						t.Error("Category should be a non-empty string")
					}
				}
			}

			// Should have valid dates
			if historyEntry.CreatedAt.IsZero() {
				t.Error("History entry created_at should be a valid date")
			}
			if historyEntry.UpdatedAt.IsZero() {
				t.Error("History entry updated_at should be a valid date")
			}

			// Should be one of: ADD, UPDATE, DELETE, NOOP
			validEvents := []Event{EventAdd, EventUpdate, EventDelete, EventNoop}
			validEvent := false
			for _, event := range validEvents {
				if historyEntry.Event == event {
					validEvent = true
					break
				}
			}
			if !validEvent {
				t.Errorf("History entry event should be one of ADD, UPDATE, DELETE, NOOP, got %s", historyEntry.Event)
			}

			// Validate conditions based on event type
			switch historyEntry.Event {
			case EventAdd:
				if historyEntry.OldMemory != nil {
					t.Error("ADD event should have null old_memory")
				}
				if historyEntry.NewMemory == nil {
					t.Error("ADD event should have non-null new_memory")
				}
			case EventUpdate:
				if historyEntry.OldMemory == nil {
					t.Error("UPDATE event should have non-null old_memory")
				}
				if historyEntry.NewMemory == nil {
					t.Error("UPDATE event should have non-null new_memory")
				}
			case EventDelete:
				if historyEntry.OldMemory == nil {
					t.Error("DELETE event should have non-null old_memory")
				}
				if historyEntry.NewMemory != nil {
					t.Error("DELETE event should have null new_memory")
				}
			}

			// Validate input messages
			if historyEntry.Input != nil {
				for _, input := range historyEntry.Input {
					// Should have string content
					if input.Content == nil {
						t.Error("Input message content should not be nil")
					}

					// Should have a role that is either 'user' or 'assistant'
					if input.Role != "user" && input.Role != "assistant" {
						t.Errorf("Input message role should be 'user' or 'assistant', got %s", input.Role)
					}
				}
			}
		}
	})

	t.Run("should delete the user successfully", func(t *testing.T) {
		ctx := context.Background()

		allUsers, err := testClient.Users(ctx)
		if err != nil {
			t.Fatalf("Failed to get users: %v", err)
		}

		var entity *User
		for _, user := range allUsers.Results {
			if user.Name == testUserID {
				entity = &user
				break
			}
		}

		if entity != nil {
			// Use the new DeleteUsers method instead of deprecated DeleteUser
			userID := entity.Name
			params := DeleteUsersParams{UserID: &userID}

			deletedUser, err := testClient.DeleteUsers(ctx, params)
			if err != nil {
				t.Fatalf("Failed to delete user: %v", err)
			}

			// Validate the deletion message
			if deletedUser.Message != "Entity deleted successfully." {
				t.Errorf("Expected 'Entity deleted successfully.', got %s", deletedUser.Message)
			}
		}
	})
}
