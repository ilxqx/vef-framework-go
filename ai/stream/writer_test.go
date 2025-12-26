package stream

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSseWriter_WriteChunk(t *testing.T) {
	tests := []struct {
		name     string
		chunk    Chunk
		expected string
	}{
		{
			name:     "SimpleChunk",
			chunk:    Chunk{"type": "test"},
			expected: `data: {"type":"test"}` + "\n\n",
		},
		{
			name:     "ChunkWithStringValue",
			chunk:    Chunk{"type": "text-delta", "delta": "Hello"},
			expected: `data: {"delta":"Hello","type":"text-delta"}` + "\n\n",
		},
		{
			name:     "ChunkWithNestedObject",
			chunk:    Chunk{"type": "data", "data": map[string]any{"key": "value"}},
			expected: `data: {"data":{"key":"value"},"type":"data"}` + "\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := newSseWriter(bufio.NewWriter(&buf))

			err := w.WriteChunk(tt.chunk)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestSseWriter_WriteDone(t *testing.T) {
	var buf bytes.Buffer
	w := newSseWriter(bufio.NewWriter(&buf))

	err := w.writeDone()

	require.NoError(t, err)
	assert.Equal(t, "data: [DONE]\n\n", buf.String())
}

func TestSseWriter_Flush(t *testing.T) {
	var buf bytes.Buffer
	bw := bufio.NewWriter(&buf)
	w := newSseWriter(bw)

	_, _ = bw.WriteString("pending data")
	err := w.Flush()

	require.NoError(t, err)
	assert.Equal(t, "pending data", buf.String())
}

func TestSseHeaders(t *testing.T) {
	assert.Equal(t, "text/event-stream", SseHeaders["Content-Type"])
	assert.Equal(t, "no-cache", SseHeaders["Cache-Control"])
	assert.Equal(t, "keep-alive", SseHeaders["Connection"])
	assert.Equal(t, "chunked", SseHeaders["Transfer-Encoding"])
	assert.Equal(t, "v1", SseHeaders["X-Vercel-AI-UI-Message-Stream"])
	assert.Equal(t, "no", SseHeaders["X-Accel-Buffering"])
}

func TestDefaultIdGenerator_Format(t *testing.T) {
	prefixes := []string{"message", "text", "reasoning", "call"}

	for _, prefix := range prefixes {
		t.Run(prefix, func(t *testing.T) {
			id := defaultIdGenerator(prefix)

			assert.True(t, strings.HasPrefix(id, prefix+"_"))
			// UUID v7 format: prefix_xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
			parts := strings.SplitN(id, "_", 2)
			require.Len(t, parts, 2)
			assert.Len(t, parts[1], 36) // UUID length
		})
	}
}

func TestDefaultIdGenerator_Uniqueness(t *testing.T) {
	ids := make(map[string]bool)

	for range 100 {
		id := defaultIdGenerator("test")
		assert.False(t, ids[id], "duplicate id generated: %s", id)
		ids[id] = true
	}
}
