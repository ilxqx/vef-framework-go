package stream

import (
	"bufio"
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/id"
)

func defaultIdGenerator(prefix string) string {
	return prefix + constants.Underscore + id.GenerateUuid()
}

type sseWriter struct {
	w *bufio.Writer
}

func newSseWriter(w *bufio.Writer) *sseWriter {
	return &sseWriter{w: w}
}

func (s *sseWriter) WriteChunk(chunk Chunk) error {
	data, err := json.Marshal(chunk)
	if err != nil {
		return fmt.Errorf("failed to marshal chunk: %w", err)
	}

	if _, err := s.w.WriteString("data: " + string(data) + "\n\n"); err != nil {
		return fmt.Errorf("failed to write sse data: %w", err)
	}

	return s.w.Flush()
}

func (s *sseWriter) Flush() error {
	return s.w.Flush()
}

func (s *sseWriter) writeDone() error {
	if _, err := s.w.WriteString("data: [DONE]\n\n"); err != nil {
		return err
	}

	return s.w.Flush()
}

// SseHeaders contains the standard headers for AI SDK UI Message Stream.
var SseHeaders = map[string]string{
	fiber.HeaderContentType:         "text/event-stream",
	fiber.HeaderCacheControl:        "no-cache",
	fiber.HeaderConnection:          "keep-alive",
	fiber.HeaderTransferEncoding:    "chunked",
	"X-Vercel-AI-UI-Message-Stream": "v1",
	"X-Accel-Buffering":             "no",
}
