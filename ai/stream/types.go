package stream

// Role represents the role of a message sender.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
	RoleSystem    Role = "system"
)

// Message represents a single message in the stream.
type Message struct {
	Role       Role
	Content    string
	ToolCalls  []ToolCall
	ToolCallID string
	Reasoning  string
	Data       map[string]any
}

// ToolCall represents a tool invocation by the AI.
type ToolCall struct {
	ID        string
	Name      string
	Arguments string
}

// Source represents a reference source (URL or document).
type Source struct {
	Type      string
	ID        string
	URL       string
	Title     string
	MediaType string
}
