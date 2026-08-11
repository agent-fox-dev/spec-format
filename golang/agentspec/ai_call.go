package agentspec

import (
	"context"
	"fmt"
	"time"
)

// CachePolicy controls prompt caching behavior for AICall.
type CachePolicy string

const (
	// CacheNone disables cache_control injection.
	CacheNone CachePolicy = "none"
	// CacheDefault injects ephemeral cache_control when the system prompt
	// exceeds the model-specific token threshold.
	CacheDefault CachePolicy = "default"
	// CacheExtended injects ephemeral cache_control with a 1-hour TTL
	// when the system prompt exceeds the model-specific token threshold.
	CacheExtended CachePolicy = "extended"
)

// Message represents a conversation message.
type Message struct {
	Role    string
	Content string
}

// SystemBlock represents a block within the system prompt,
// potentially annotated with cache_control metadata.
type SystemBlock struct {
	Type         string
	Text         string
	CacheControl map[string]string
}

// Tool represents an Anthropic tool definition.
type Tool struct {
	Name        string
	Description string
	InputSchema any
}

// ContentBlock represents a single content block in an LLM response.
type ContentBlock struct {
	Type  string
	Text  string
	ID    string
	Name  string
	Input any
}

// MessageRequest holds the parameters for a single LLM API call.
type MessageRequest struct {
	Model       string
	MaxTokens   int
	Messages    []Message
	System      []SystemBlock
	Tools       []Tool
	ToolChoice  any
	Temperature *float64
}

// MessageResponse holds the result of a single LLM API call.
type MessageResponse struct {
	Content    []ContentBlock
	StopReason string
}

// Doer is the interface for sending LLM requests. AICall uses this
// internally; tests can provide a mock implementation via AICallOptions.Doer.
type Doer interface {
	CreateMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error)
}

// APIError represents an HTTP API error with a status code.
// AICall treats 429 and 5xx status codes as retryable.
type APIError struct {
	StatusCode int
	Msg        string
}

func (e *APIError) Error() string { return e.Msg }

// IsRetryable reports whether the error is a transient error that
// should be retried (rate limit 429, server 5xx, or connection error).
func IsRetryable(err error) bool {
	// TODO: implement
	return false
}

// AICallOptions configures an AICall invocation.
type AICallOptions struct {
	// ModelTier is the model tier or model ID to use.
	ModelTier string
	// MaxTokens is the maximum number of tokens to generate.
	// Defaults to 65536 when zero.
	MaxTokens int
	// Messages is the conversation message history.
	Messages []Message
	// System is the system prompt text.
	System string
	// Context describes the call site for logging/error messages.
	Context string
	// CachePolicy controls prompt caching behavior.
	// Defaults to CacheDefault when empty.
	CachePolicy CachePolicy
	// Temperature controls response randomness.
	Temperature *float64
	// Tools is the list of tool definitions available to the model.
	Tools []Tool
	// ToolChoice constrains which tool the model should use.
	ToolChoice any

	// Doer overrides the default Anthropic client for testing.
	// When nil, AICall creates a platform-aware client based on
	// environment variables.
	Doer Doer

	// ClientFactory overrides client creation for testing.
	// Called with the detected platform string ("direct", "vertex",
	// "bedrock"). When nil and Doer is nil, AICall creates the
	// appropriate Anthropic SDK client.
	ClientFactory func(platform string) (Doer, error)

	// SleepFunc overrides time.Sleep for testing retry delays.
	// When nil, AICall uses time.Sleep.
	SleepFunc func(time.Duration)
}

// ApplyDefaults fills in default values for unset AICallOptions fields.
// MaxTokens defaults to 65536 and CachePolicy defaults to CacheDefault.
func ApplyDefaults(opts *AICallOptions) {
	// TODO: implement defaults logic
}

// AICall is the central LLM interface function. It resolves the model tier,
// creates a platform-aware Anthropic client, streams the response with retry
// and prompt caching, and returns the complete response text.
//
// Returns (responseText, rawResponse, nil) on success, or ("", nil, error)
// on failure. The error is wrapped as an *AgentError when appropriate.
func AICall(ctx context.Context, opts AICallOptions) (string, any, error) {
	// TODO: implement
	return "", nil, fmt.Errorf("AICall: not implemented")
}

// RetryDelays are the exponential backoff delays between retry attempts.
var RetryDelays = []time.Duration{
	2 * time.Second,
	30 * time.Second,
	60 * time.Second,
}
