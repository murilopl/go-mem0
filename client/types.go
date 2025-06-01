package client

import "time"

// API version and output format enums
type APIVersion string
type OutputFormat string
type Feedback string
type Event string
type WebhookEvent string

const (
	APIVersionV1 APIVersion = "v1"
	APIVersionV2 APIVersion = "v2"

	OutputFormatV1   OutputFormat = "v1.0"
	OutputFormatV1_1 OutputFormat = "v1.1"

	FeedbackPositive     Feedback = "POSITIVE"
	FeedbackNegative     Feedback = "NEGATIVE"
	FeedbackVeryNegative Feedback = "VERY_NEGATIVE"

	EventAdd    Event = "ADD"
	EventUpdate Event = "UPDATE"
	EventDelete Event = "DELETE"
	EventNoop   Event = "NOOP"

	WebhookEventMemoryAdded   WebhookEvent = "memory_add"
	WebhookEventMemoryUpdated WebhookEvent = "memory_update"
	WebhookEventMemoryDeleted WebhookEvent = "memory_delete"
)

// MultiModalMessages represents image content in messages
type MultiModalMessages struct {
	Type     string `json:"type"`
	ImageURL struct {
		URL string `json:"url"`
	} `json:"image_url"`
}

// Message represents a chat message
type Message struct {
	Role    string      `json:"role"` // "user" or "assistant"
	Content interface{} `json:"content"` // string or MultiModalMessages
}

// MemoryOptions contains options for memory operations
type MemoryOptions struct {
	APIVersion         *APIVersion           `json:"api_version,omitempty"`
	Version            *APIVersion           `json:"version,omitempty"`
	UserID             *string               `json:"user_id,omitempty"`
	AgentID            *string               `json:"agent_id,omitempty"`
	AppID              *string               `json:"app_id,omitempty"`
	RunID              *string               `json:"run_id,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	Filters            map[string]interface{} `json:"filters,omitempty"`
	OrgName            *string               `json:"org_name,omitempty"` // Deprecated
	ProjectName        *string               `json:"project_name,omitempty"` // Deprecated
	OrgID              interface{}           `json:"org_id,omitempty"` // string or number
	ProjectID          interface{}           `json:"project_id,omitempty"` // string or number
	Infer              *bool                 `json:"infer,omitempty"`
	Page               *int                  `json:"page,omitempty"`
	PageSize           *int                  `json:"page_size,omitempty"`
	Includes           *string               `json:"includes,omitempty"`
	Excludes           *string               `json:"excludes,omitempty"`
	EnableGraph        *bool                 `json:"enable_graph,omitempty"`
	StartDate          *string               `json:"start_date,omitempty"`
	EndDate            *string               `json:"end_date,omitempty"`
	CustomCategories   []map[string]interface{} `json:"custom_categories,omitempty"`
	CustomInstructions *string               `json:"custom_instructions,omitempty"`
	Timestamp          *int64                `json:"timestamp,omitempty"`
	OutputFormat       *OutputFormat         `json:"output_format,omitempty"`
	AsyncMode          *bool                 `json:"async_mode,omitempty"`
}

// SearchOptions extends MemoryOptions with search-specific fields
type SearchOptions struct {
	MemoryOptions
	Limit                      *int      `json:"limit,omitempty"`
	EnableGraph                *bool     `json:"enable_graph,omitempty"`
	Threshold                  *float64  `json:"threshold,omitempty"`
	TopK                       *int      `json:"top_k,omitempty"`
	OnlyMetadataBasedSearch    *bool     `json:"only_metadata_based_search,omitempty"`
	KeywordSearch              *bool     `json:"keyword_search,omitempty"`
	Fields                     []string  `json:"fields,omitempty"`
	Categories                 []string  `json:"categories,omitempty"`
	Rerank                     *bool     `json:"rerank,omitempty"`
}

// ProjectOptions contains options for project operations
type ProjectOptions struct {
	Fields []string `json:"fields,omitempty"`
}

// MemoryData represents memory content
type MemoryData struct {
	Memory string `json:"memory"`
}

// Memory represents a memory object
type Memory struct {
	ID         string       `json:"id"`
	Messages   []Message    `json:"messages,omitempty"`
	Event      *Event       `json:"event,omitempty"`
	Data       *MemoryData  `json:"data,omitempty"`
	Memory     *string      `json:"memory,omitempty"`
	UserID     *string      `json:"user_id,omitempty"`
	Hash       *string      `json:"hash,omitempty"`
	Categories []string     `json:"categories,omitempty"`
	CreatedAt  *time.Time   `json:"created_at,omitempty"`
	UpdatedAt  *time.Time   `json:"updated_at,omitempty"`
	MemoryType *string      `json:"memory_type,omitempty"`
	Score      *float64     `json:"score,omitempty"`
	Metadata   interface{}  `json:"metadata,omitempty"`
	Owner      *string      `json:"owner,omitempty"`
	AgentID    *string      `json:"agent_id,omitempty"`
	AppID      *string      `json:"app_id,omitempty"`
	RunID      *string      `json:"run_id,omitempty"`
}

// MemoryHistory represents memory change history
type MemoryHistory struct {
	ID        string    `json:"id"`
	MemoryID  string    `json:"memory_id"`
	Input     []Message `json:"input"`
	OldMemory *string   `json:"old_memory"`
	NewMemory *string   `json:"new_memory"`
	UserID    string    `json:"user_id"`
	Categories []string `json:"categories"`
	Event     Event     `json:"event"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MemoryUpdateBody represents data for batch memory updates
type MemoryUpdateBody struct {
	MemoryID string `json:"memoryId"`
	Text     string `json:"text"`
}

// User represents a user entity
type User struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	TotalMemories int       `json:"total_memories"`
	Owner         string    `json:"owner"`
	Type          string    `json:"type"`
}

// AllUsers represents paginated user results
type AllUsers struct {
	Count    int           `json:"count"`
	Results  []User        `json:"results"`
	Next     interface{}   `json:"next"`
	Previous interface{}   `json:"previous"`
}

// ProjectResponse represents project data
type ProjectResponse struct {
	CustomInstructions *string                  `json:"custom_instructions,omitempty"`
	CustomCategories   []string                 `json:"custom_categories,omitempty"`
	Additional         map[string]interface{}   `json:"-"` // For other fields
}

// PromptUpdatePayload represents data for updating project prompts
type PromptUpdatePayload struct {
	CustomInstructions *string                    `json:"custom_instructions,omitempty"`
	CustomCategories   []map[string]interface{}   `json:"custom_categories,omitempty"`
	Additional         map[string]interface{}     `json:"-"` // For other fields
}

// Webhook represents a webhook configuration
type Webhook struct {
	WebhookID  *string        `json:"webhook_id,omitempty"`
	Name       string         `json:"name"`
	URL        string         `json:"url"`
	Project    *string        `json:"project,omitempty"`
	CreatedAt  *time.Time     `json:"created_at,omitempty"`
	UpdatedAt  *time.Time     `json:"updated_at,omitempty"`
	IsActive   *bool          `json:"is_active,omitempty"`
	EventTypes []WebhookEvent `json:"event_types,omitempty"`
}

// WebhookPayload represents data for webhook operations
type WebhookPayload struct {
	EventTypes []WebhookEvent `json:"eventTypes"`
	ProjectID  string         `json:"projectId"`
	WebhookID  string         `json:"webhookId"`
	Name       string         `json:"name"`
	URL        string         `json:"url"`
}

// FeedbackPayload represents feedback data
type FeedbackPayload struct {
	MemoryID       string    `json:"memory_id"`
	Feedback       *Feedback `json:"feedback,omitempty"`
	FeedbackReason *string   `json:"feedback_reason,omitempty"`
}

// DeleteUsersParams represents parameters for deleting users
type DeleteUsersParams struct {
	UserID  *string `json:"user_id,omitempty"`
	AgentID *string `json:"agent_id,omitempty"`
	AppID   *string `json:"app_id,omitempty"`
	RunID   *string `json:"run_id,omitempty"`
}

// DeleteUserData represents deprecated user deletion data
type DeleteUserData struct {
	EntityID   int    `json:"entity_id"`
	EntityType string `json:"entity_type"`
}

// DeleteWebhookData represents webhook deletion data
type DeleteWebhookData struct {
	WebhookID string `json:"webhookId"`
}

// Generic response types
type MessageResponse struct {
	Message string `json:"message"`
}

type PingResponse struct {
	Status     string `json:"status"`
	OrgID      string `json:"org_id,omitempty"`
	ProjectID  string `json:"project_id,omitempty"`
	UserEmail  string `json:"user_email,omitempty"`
	Message    string `json:"message,omitempty"`
}