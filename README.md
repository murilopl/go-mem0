# Go Mem0 Client

[![CI](https://github.com/murilopl/go-mem0/workflows/CI/badge.svg)](https://github.com/murilopl/go-mem0/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/murilopl/go-mem0)](https://goreportcard.com/report/github.com/murilopl/go-mem0)
[![GoDoc](https://godoc.org/github.com/murilopl/go-mem0/client?status.svg)](https://godoc.org/github.com/murilopl/go-mem0/client)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A Go client library for the [Mem0](https://mem0.ai/) API - an intelligent memory layer for AI applications.

**Note**: This client is designed for the Mem0 managed service and does not support self-hosted Mem0 instances.

## Installation

```bash
go get github.com/murilopl/go-mem0
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/murilopl/go-mem0/client"
)

func main() {
    // Create a new memory client
    memoryClient, err := client.NewMemoryClient(client.ClientOptions{
        APIKey: "your-mem0-api-key",
    })
    if err != nil {
        log.Fatalf("Failed to create memory client: %v", err)
    }

    ctx := context.Background()

    // Add memories with user context
    messages := []client.Message{
        {Role: "user", Content: "My name is John and I love programming in Go"},
        {Role: "assistant", Content: "Nice to meet you John! Go is a great language."},
    }

    userID := "user-123"
    options := client.MemoryOptions{UserID: &userID}
    memories, err := memoryClient.Add(ctx, messages, options)
    if err != nil {
        log.Fatalf("Failed to add memories: %v", err)
    }
    fmt.Printf("Added %d memories\n", len(memories))

    // Search memories
    searchOptions := client.SearchOptions{
        MemoryOptions: client.MemoryOptions{UserID: &userID},
    }
    searchResults, err := memoryClient.Search(ctx, "programming languages", searchOptions)
    if err != nil {
        log.Fatalf("Failed to search memories: %v", err)
    }
    fmt.Printf("Found %d memories\n", len(searchResults))

    // Get all memories
    getAllOptions := client.SearchOptions{
        MemoryOptions: client.MemoryOptions{UserID: &userID},
    }
    allMemories, err := memoryClient.GetAll(ctx, getAllOptions)
    if err != nil {
        log.Fatalf("Failed to get all memories: %v", err)
    }
    fmt.Printf("Total memories: %d\n", len(allMemories))
}
```

## Features

### ✅ Core Memory Operations
- **Add**: Store conversations and experiences as memories
- **Get**: Retrieve specific memories by ID
- **GetAll**: List all memories with filtering options
- **Search**: Search memories using natural language queries
- **Update**: Modify existing memories
- **Delete**: Remove specific memories
- **DeleteAll**: Bulk delete memories with filters

### ✅ Advanced Features
- **Batch Operations**: Bulk update and delete multiple memories
- **Memory History**: Track changes to memories over time
- **User Management**: Manage users and entities
- **API Versioning**: Support for both v1 and v2 endpoints
- **Context Support**: All methods accept context for cancellation/timeouts

### 🔧 Configuration Options
- **Organizations & Projects**: Support for multi-tenant setups
- **Custom Hosts**: Use custom API endpoints
- **Metadata & Filters**: Rich filtering and tagging capabilities
- **Search Options**: Advanced search with thresholds, categories, and more

## API Reference

### Authentication

To use the Mem0 Go client, you need an API key from the [Mem0 managed service](https://app.mem0.ai/dashboard/api-keys). The client connects to `https://api.mem0.ai` by default and accepts the API key directly in the constructor:

```go
client, err := client.NewMemoryClient(client.ClientOptions{
    APIKey: "m0-your-api-key-here",
})
```

This client is specifically designed for the Mem0 managed service and does not support self-hosted Mem0 instances.

### Client Initialization

```go
// Basic initialization
client, err := client.NewMemoryClient(client.ClientOptions{
    APIKey: "your-mem0-api-key",
})

// Advanced initialization with custom configuration
host := "https://api.mem0.ai"
orgID := "your-org-id"
projectID := "your-project-id"

client, err := client.NewMemoryClient(client.ClientOptions{
    APIKey:           "your-mem0-api-key",
    Host:             &host,         // Optional: custom API host
    OrganizationID:   &orgID,        // Optional: organization context
    ProjectID:        &projectID,    // Optional: project context
})
```

### Memory Operations

#### Add Memories
```go
messages := []client.Message{
    {Role: "user", Content: "Hello, I'm learning Go"},
    {Role: "assistant", Content: "Great! Go is an excellent language"},
}

userID := "user-123"
options := client.MemoryOptions{
    UserID: &userID,
    Metadata: map[string]interface{}{
        "session": "learning-session-1",
        "topic": "programming",
    },
}

memories, err := client.Add(ctx, messages, options)
```

#### Search Memories
```go
// Simple search
results, err := client.Search(ctx, "What programming languages do I know?")

// Advanced search with options
threshold := 0.7
apiVersion := client.APIVersionV2
options := client.SearchOptions{
    MemoryOptions: client.MemoryOptions{
        UserID: &userID,
        APIVersion: &apiVersion,
    },
    Threshold: &threshold,
    Limit: &limit,
}
results, err := client.Search(ctx, "programming", options)
```

#### Get All Memories
```go
// Get all memories for a user
options := client.SearchOptions{
    MemoryOptions: client.MemoryOptions{
        UserID: &userID,
    },
}
memories, err := client.GetAll(ctx, options)

// With pagination
page := 1
pageSize := 10
options.Page = &page
options.PageSize = &pageSize
memories, err := client.GetAll(ctx, options)
```

### Batch Operations

```go
// Batch update
updates := []client.MemoryUpdateBody{
    {MemoryID: "mem-1", Text: "Updated content 1"},
    {MemoryID: "mem-2", Text: "Updated content 2"},
}
result, err := client.BatchUpdate(ctx, updates)

// Batch delete
memoryIDs := []string{"mem-1", "mem-2", "mem-3"}
result, err := client.BatchDelete(ctx, memoryIDs)
```

### User Management

```go
// Get all users
users, err := client.Users(ctx)

// Delete specific user
params := client.DeleteUsersParams{
    UserID: &userID,
}
result, err := client.DeleteUsers(ctx, params)

// Delete all users
result, err := client.DeleteUsers(ctx)
```

### Memory History

```go
// Get history of changes for a memory
history, err := client.History(ctx, memoryID)

for _, entry := range history {
    fmt.Printf("Event: %s, Old: %s, New: %s\n", 
        entry.Event, 
        stringValue(entry.OldMemory), 
        stringValue(entry.NewMemory))
}
```

## Error Handling

The client provides structured error types:

```go
memories, err := client.Add(ctx, messages)
if err != nil {
    switch e := err.(type) {
    case *client.APIError:
        fmt.Printf("API Error %d: %s\n", e.StatusCode, e.Message)
    case *client.ValidationError:
        fmt.Printf("Validation Error in %s: %s\n", e.Field, e.Message)
    default:
        fmt.Printf("Unknown error: %v\n", err)
    }
    return
}
```

## Testing

Run the test suite:

```bash
# Unit tests only
go test -short ./client

# Integration tests (requires API key)
go test ./client
```

## Type Definitions

The client includes comprehensive type definitions for all API objects:

- `Memory`: Individual memory objects with metadata
- `Message`: Chat messages with role and content
- `MemoryOptions`: Configuration for memory operations
- `SearchOptions`: Advanced search parameters
- `User`: User/entity information
- `MemoryHistory`: Memory change tracking
- And many more...

## API Versions

The client supports both Mem0 API versions:

- **v1 (default)**: Stable API with core functionality
- **v2**: Enhanced API with advanced features

```go
apiVersion := client.APIVersionV2
options := client.MemoryOptions{
    APIVersion: &apiVersion,
}
```

## License

This project follows the same license as the original Mem0 project.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## Related Projects

- [Mem0 Python SDK](https://github.com/mem0ai/mem0) - Official Python client
- [Mem0 TypeScript SDK](../ts-client/) - TypeScript client (converted from)