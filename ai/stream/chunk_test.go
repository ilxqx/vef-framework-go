package stream

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStartChunk(t *testing.T) {
	chunk := NewStartChunk("msg_123")

	assert.Equal(t, ChunkTypeStart, chunk["type"])
	assert.Equal(t, "msg_123", chunk["messageId"])
}

func TestNewFinishChunk(t *testing.T) {
	chunk := NewFinishChunk()

	assert.Equal(t, ChunkTypeFinish, chunk["type"])
	assert.Len(t, chunk, 1)
}

func TestNewStartStepChunk(t *testing.T) {
	chunk := NewStartStepChunk()

	assert.Equal(t, ChunkTypeStartStep, chunk["type"])
	assert.Len(t, chunk, 1)
}

func TestNewFinishStepChunk(t *testing.T) {
	chunk := NewFinishStepChunk()

	assert.Equal(t, ChunkTypeFinishStep, chunk["type"])
	assert.Len(t, chunk, 1)
}

func TestNewErrorChunk(t *testing.T) {
	chunk := NewErrorChunk("something went wrong")

	assert.Equal(t, ChunkTypeError, chunk["type"])
	assert.Equal(t, "something went wrong", chunk["errorText"])
}

func TestTextChunks(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() Chunk
		expected Chunk
	}{
		{
			name: "TextStart",
			fn:   func() Chunk { return NewTextStartChunk("text_1") },
			expected: Chunk{
				"type": ChunkTypeTextStart,
				"id":   "text_1",
			},
		},
		{
			name: "TextDelta",
			fn:   func() Chunk { return NewTextDeltaChunk("text_1", "Hello") },
			expected: Chunk{
				"type":  ChunkTypeTextDelta,
				"id":    "text_1",
				"delta": "Hello",
			},
		},
		{
			name: "TextEnd",
			fn:   func() Chunk { return NewTextEndChunk("text_1") },
			expected: Chunk{
				"type": ChunkTypeTextEnd,
				"id":   "text_1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := tt.fn()
			assert.Equal(t, tt.expected, chunk)
		})
	}
}

func TestReasoningChunks(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() Chunk
		expected Chunk
	}{
		{
			name: "ReasoningStart",
			fn:   func() Chunk { return NewReasoningStartChunk("reasoning_1") },
			expected: Chunk{
				"type": ChunkTypeReasoningStart,
				"id":   "reasoning_1",
			},
		},
		{
			name: "ReasoningDelta",
			fn:   func() Chunk { return NewReasoningDeltaChunk("reasoning_1", "thinking...") },
			expected: Chunk{
				"type":  ChunkTypeReasoningDelta,
				"id":    "reasoning_1",
				"delta": "thinking...",
			},
		},
		{
			name: "ReasoningEnd",
			fn:   func() Chunk { return NewReasoningEndChunk("reasoning_1") },
			expected: Chunk{
				"type": ChunkTypeReasoningEnd,
				"id":   "reasoning_1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := tt.fn()
			assert.Equal(t, tt.expected, chunk)
		})
	}
}

func TestToolChunks(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() Chunk
		expected Chunk
	}{
		{
			name: "ToolInputStart",
			fn:   func() Chunk { return NewToolInputStartChunk("call_1", "get_weather") },
			expected: Chunk{
				"type":       ChunkTypeToolInputStart,
				"toolCallId": "call_1",
				"toolName":   "get_weather",
			},
		},
		{
			name: "ToolInputDelta",
			fn:   func() Chunk { return NewToolInputDeltaChunk("call_1", `{"city":`) },
			expected: Chunk{
				"type":           ChunkTypeToolInputDelta,
				"toolCallId":     "call_1",
				"inputTextDelta": `{"city":`,
			},
		},
		{
			name: "ToolInputAvailable",
			fn: func() Chunk {
				return NewToolInputAvailableChunk("call_1", "get_weather", map[string]string{"city": "Beijing"})
			},
			expected: Chunk{
				"type":       ChunkTypeToolInputAvailable,
				"toolCallId": "call_1",
				"toolName":   "get_weather",
				"input":      map[string]string{"city": "Beijing"},
			},
		},
		{
			name: "ToolOutputAvailable",
			fn: func() Chunk {
				return NewToolOutputAvailableChunk("call_1", map[string]any{"temp": 25, "unit": "celsius"})
			},
			expected: Chunk{
				"type":       ChunkTypeToolOutputAvailable,
				"toolCallId": "call_1",
				"output":     map[string]any{"temp": 25, "unit": "celsius"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := tt.fn()
			assert.Equal(t, tt.expected, chunk)
		})
	}
}

func TestSourceChunks(t *testing.T) {
	t.Run("SourceUrlWithTitle", func(t *testing.T) {
		chunk := NewSourceUrlChunk("src_1", "https://example.com", "Example Site")

		assert.Equal(t, ChunkTypeSourceUrl, chunk["type"])
		assert.Equal(t, "src_1", chunk["sourceId"])
		assert.Equal(t, "https://example.com", chunk["url"])
		assert.Equal(t, "Example Site", chunk["title"])
	})

	t.Run("SourceUrlWithoutTitle", func(t *testing.T) {
		chunk := NewSourceUrlChunk("src_1", "https://example.com", "")

		assert.Equal(t, ChunkTypeSourceUrl, chunk["type"])
		assert.Equal(t, "src_1", chunk["sourceId"])
		assert.Equal(t, "https://example.com", chunk["url"])
		assert.NotContains(t, chunk, "title")
	})

	t.Run("SourceDocumentWithTitle", func(t *testing.T) {
		chunk := NewSourceDocumentChunk("src_2", "application/pdf", "Report.pdf")

		assert.Equal(t, ChunkTypeSourceDocument, chunk["type"])
		assert.Equal(t, "src_2", chunk["sourceId"])
		assert.Equal(t, "application/pdf", chunk["mediaType"])
		assert.Equal(t, "Report.pdf", chunk["title"])
	})

	t.Run("SourceDocumentWithoutTitle", func(t *testing.T) {
		chunk := NewSourceDocumentChunk("src_2", "application/pdf", "")

		assert.Equal(t, ChunkTypeSourceDocument, chunk["type"])
		assert.Equal(t, "src_2", chunk["sourceId"])
		assert.Equal(t, "application/pdf", chunk["mediaType"])
		assert.NotContains(t, chunk, "title")
	})
}

func TestNewFileChunk(t *testing.T) {
	chunk := NewFileChunk("file_1", "image/png", "https://cdn.example.com/image.png")

	assert.Equal(t, ChunkTypeFile, chunk["type"])
	assert.Equal(t, "file_1", chunk["fileId"])
	assert.Equal(t, "image/png", chunk["mediaType"])
	assert.Equal(t, "https://cdn.example.com/image.png", chunk["url"])
}

func TestNewDataChunk(t *testing.T) {
	tests := []struct {
		name     string
		dataType string
		data     any
	}{
		{
			name:     "StringData",
			dataType: "status",
			data:     "processing",
		},
		{
			name:     "MapData",
			dataType: "metadata",
			data:     map[string]any{"key": "value"},
		},
		{
			name:     "SliceData",
			dataType: "items",
			data:     []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := NewDataChunk(tt.dataType, tt.data)

			assert.Equal(t, ChunkType("data-"+tt.dataType), chunk["type"])
			assert.Equal(t, tt.data, chunk["data"])
		})
	}
}
