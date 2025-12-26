package stream

import "io"

// MessageSource produces streaming messages. Returns io.EOF when complete.
type MessageSource interface {
	Recv() (Message, error)
	Close() error
}

// StreamWriter writes UI message stream chunks.
type StreamWriter interface {
	WriteChunk(chunk Chunk) error
	Flush() error
}

// CallbackWriter provides methods to push messages in callback-based sources.
type CallbackWriter interface {
	WriteText(content string)
	WriteToolCall(id, name, arguments string)
	WriteToolResult(toolCallId, content string)
	WriteReasoning(reasoning string)
	WriteData(dataType string, data any)
	WriteMessage(msg Message)
}

// ResponseWriter is compatible with fiber.Ctx.SendStreamWriter.
type ResponseWriter interface {
	io.Writer
}
