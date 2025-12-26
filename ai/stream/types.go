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
	ToolCallId string
	Reasoning  string
	Data       map[string]any
}

// ToolCall represents a tool invocation by the AI.
type ToolCall struct {
	Id        string
	Name      string
	Arguments string
}

// Source represents a reference source (url or document).
type Source struct {
	Type      string
	Id        string
	Url       string
	Title     string
	MediaType string
}
