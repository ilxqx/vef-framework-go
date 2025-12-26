package stream

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/constants"
)

// Builder provides a fluent interface for building UI message streams.
type Builder struct {
	opts      Options
	source    MessageSource
	messageId string
	headers   map[string]string
}

// New creates a new stream builder with default options.
func New() *Builder {
	return &Builder{
		opts:    DefaultOptions(),
		headers: make(map[string]string),
	}
}

func (b *Builder) WithSource(source MessageSource) *Builder {
	b.source = source

	return b
}

func (b *Builder) WithMessageId(id string) *Builder {
	b.messageId = id

	return b
}

func (b *Builder) WithReasoning(enabled bool) *Builder {
	b.opts.SendReasoning = enabled

	return b
}

func (b *Builder) WithSources(enabled bool) *Builder {
	b.opts.SendSources = enabled

	return b
}

func (b *Builder) WithStart(enabled bool) *Builder {
	b.opts.SendStart = enabled

	return b
}

func (b *Builder) WithFinish(enabled bool) *Builder {
	b.opts.SendFinish = enabled

	return b
}

func (b *Builder) OnError(handler func(err error) string) *Builder {
	b.opts.OnError = handler

	return b
}

func (b *Builder) OnFinish(handler func(content string)) *Builder {
	b.opts.OnFinish = handler

	return b
}

func (b *Builder) WithIdGenerator(gen func(prefix string) string) *Builder {
	b.opts.GenerateId = gen

	return b
}

func (b *Builder) WithHeader(key, value string) *Builder {
	b.headers[key] = value

	return b
}

// Stream executes the stream and writes to a Fiber context.
func (b *Builder) Stream(ctx fiber.Ctx) error {
	if b.source == nil {
		return ErrSourceRequired
	}

	for k, v := range SseHeaders {
		ctx.Set(k, v)
	}

	for k, v := range b.headers {
		ctx.Set(k, v)
	}

	return ctx.SendStreamWriter(func(w *bufio.Writer) {
		b.streamToWriter(w)
	})
}

// StreamToWriter streams messages to a bufio.Writer.
func (b *Builder) StreamToWriter(w *bufio.Writer) {
	b.streamToWriter(w)
}

func (b *Builder) streamToWriter(w *bufio.Writer) {
	defer func() { _ = b.source.Close() }()

	writer := newSseWriter(w)

	generateId := b.opts.GenerateId
	if generateId == nil {
		generateId = defaultIdGenerator
	}

	messageId := b.messageId
	if messageId == constants.Empty {
		messageId = generateId("message")
	}

	textId := generateId("text")
	reasoningId := generateId("reasoning")
	textStarted := false
	reasoningStarted := false

	var fullContent string

	if b.opts.SendStart {
		_ = writer.WriteChunk(NewStartChunk(messageId))
		_ = writer.WriteChunk(NewStartStepChunk())
	}

	for {
		msg, err := b.source.Recv()
		if errors.Is(err, io.EOF) {
			if textStarted {
				_ = writer.WriteChunk(NewTextEndChunk(textId))
			}

			if reasoningStarted {
				_ = writer.WriteChunk(NewReasoningEndChunk(reasoningId))
			}

			if b.opts.SendFinish {
				_ = writer.WriteChunk(NewFinishStepChunk())
				_ = writer.WriteChunk(NewFinishChunk())
			}

			_ = writer.writeDone()

			if b.opts.OnFinish != nil {
				b.opts.OnFinish(fullContent)
			}

			return
		}

		if err != nil {
			errorText := err.Error()
			if b.opts.OnError != nil {
				errorText = b.opts.OnError(err)
			}

			_ = writer.WriteChunk(NewErrorChunk(errorText))
			_ = writer.writeDone()

			return
		}

		// Handle reasoning
		if b.opts.SendReasoning && msg.Reasoning != constants.Empty {
			if !reasoningStarted {
				_ = writer.WriteChunk(NewReasoningStartChunk(reasoningId))
				reasoningStarted = true
			}

			_ = writer.WriteChunk(NewReasoningDeltaChunk(reasoningId, msg.Reasoning))
		}

		// Handle tool calls
		for _, tc := range msg.ToolCalls {
			toolCallId := tc.Id
			if toolCallId == constants.Empty {
				toolCallId = generateId("call")
			}

			_ = writer.WriteChunk(NewToolInputStartChunk(toolCallId, tc.Name))

			var input any
			if err := json.Unmarshal([]byte(tc.Arguments), &input); err != nil {
				input = tc.Arguments
			}

			_ = writer.WriteChunk(NewToolInputAvailableChunk(toolCallId, tc.Name, input))
		}

		// Handle tool results
		if msg.Role == RoleTool && msg.ToolCallId != constants.Empty {
			var output any
			if err := json.Unmarshal([]byte(msg.Content), &output); err != nil {
				output = msg.Content
			}

			_ = writer.WriteChunk(NewToolOutputAvailableChunk(msg.ToolCallId, output))

			continue
		}

		// Handle custom data
		for dataType, data := range msg.Data {
			_ = writer.WriteChunk(NewDataChunk(dataType, data))
		}

		// Handle text content
		if msg.Content != constants.Empty {
			if reasoningStarted {
				_ = writer.WriteChunk(NewReasoningEndChunk(reasoningId))
				reasoningStarted = false
				reasoningId = generateId("reasoning")
			}

			if !textStarted {
				_ = writer.WriteChunk(NewTextStartChunk(textId))
				textStarted = true
			}

			_ = writer.WriteChunk(NewTextDeltaChunk(textId, msg.Content))
			fullContent += msg.Content
		}
	}
}
