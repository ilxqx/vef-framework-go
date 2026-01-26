package stream

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChannelSource(t *testing.T) {
	t.Run("ReceivesMessagesUntilChannelClosed", func(t *testing.T) {
		ch := make(chan Message, 3)
		ch <- Message{Role: RoleUser, Content: "Hello"}

		ch <- Message{Role: RoleAssistant, Content: "Hi"}

		ch <- Message{Role: RoleAssistant, Content: "there"}

		close(ch)

		source := NewChannelSource(ch)
		defer source.Close()

		msg1, err := source.Recv()
		require.NoError(t, err)
		assert.Equal(t, RoleUser, msg1.Role)
		assert.Equal(t, "Hello", msg1.Content)

		msg2, err := source.Recv()
		require.NoError(t, err)
		assert.Equal(t, RoleAssistant, msg2.Role)
		assert.Equal(t, "Hi", msg2.Content)

		msg3, err := source.Recv()
		require.NoError(t, err)
		assert.Equal(t, "there", msg3.Content)

		_, err = source.Recv()
		assert.ErrorIs(t, err, io.EOF)
	})

	t.Run("ReturnsEofAfterClose", func(t *testing.T) {
		ch := make(chan Message, 1)
		ch <- Message{Role: RoleUser, Content: "test"}

		source := NewChannelSource(ch)
		err := source.Close()
		require.NoError(t, err)

		_, err = source.Recv()
		assert.ErrorIs(t, err, io.EOF)
	})

	t.Run("HandlesEmptyChannel", func(t *testing.T) {
		ch := make(chan Message)
		close(ch)

		source := NewChannelSource(ch)
		defer source.Close()

		_, err := source.Recv()
		assert.ErrorIs(t, err, io.EOF)
	})
}

func TestCallbackSource(t *testing.T) {
	t.Run("ReceivesTextMessages", func(t *testing.T) {
		source := NewCallbackSource(func(w CallbackWriter) error {
			w.WriteText("Hello")
			w.WriteText(" World")

			return nil
		})
		defer source.Close()

		msg1, err := source.Recv()
		require.NoError(t, err)
		assert.Equal(t, RoleAssistant, msg1.Role)
		assert.Equal(t, "Hello", msg1.Content)

		msg2, err := source.Recv()
		require.NoError(t, err)
		assert.Equal(t, " World", msg2.Content)

		_, err = source.Recv()
		assert.ErrorIs(t, err, io.EOF)
	})

	t.Run("ReceivesToolCalls", func(t *testing.T) {
		source := NewCallbackSource(func(w CallbackWriter) error {
			w.WriteToolCall("call_1", "get_weather", `{"city":"Beijing"}`)

			return nil
		})
		defer source.Close()

		msg, err := source.Recv()
		require.NoError(t, err)
		assert.Equal(t, RoleAssistant, msg.Role)
		require.Len(t, msg.ToolCalls, 1)
		assert.Equal(t, "call_1", msg.ToolCalls[0].ID)
		assert.Equal(t, "get_weather", msg.ToolCalls[0].Name)
		assert.Equal(t, `{"city":"Beijing"}`, msg.ToolCalls[0].Arguments)
	})

	t.Run("ReceivesToolResults", func(t *testing.T) {
		source := NewCallbackSource(func(w CallbackWriter) error {
			w.WriteToolResult("call_1", `{"temp":25}`)

			return nil
		})
		defer source.Close()

		msg, err := source.Recv()
		require.NoError(t, err)
		assert.Equal(t, RoleTool, msg.Role)
		assert.Equal(t, "call_1", msg.ToolCallID)
		assert.Equal(t, `{"temp":25}`, msg.Content)
	})

	t.Run("ReceivesReasoning", func(t *testing.T) {
		source := NewCallbackSource(func(w CallbackWriter) error {
			w.WriteReasoning("Let me think...")

			return nil
		})
		defer source.Close()

		msg, err := source.Recv()
		require.NoError(t, err)
		assert.Equal(t, RoleAssistant, msg.Role)
		assert.Equal(t, "Let me think...", msg.Reasoning)
	})

	t.Run("ReceivesCustomData", func(t *testing.T) {
		source := NewCallbackSource(func(w CallbackWriter) error {
			w.WriteData("status", map[string]any{"progress": 50})

			return nil
		})
		defer source.Close()

		msg, err := source.Recv()
		require.NoError(t, err)
		assert.Equal(t, RoleAssistant, msg.Role)
		assert.Equal(t, map[string]any{"progress": 50}, msg.Data["status"])
	})

	t.Run("ReceivesFullMessage", func(t *testing.T) {
		customMsg := Message{
			Role:    RoleSystem,
			Content: "System prompt",
		}

		source := NewCallbackSource(func(w CallbackWriter) error {
			w.WriteMessage(customMsg)

			return nil
		})
		defer source.Close()

		msg, err := source.Recv()
		require.NoError(t, err)
		assert.Equal(t, customMsg, msg)
	})

	t.Run("PropagatesError", func(t *testing.T) {
		expectedErr := io.ErrUnexpectedEOF

		source := NewCallbackSource(func(w CallbackWriter) error {
			w.WriteText("partial")

			return expectedErr
		})
		defer source.Close()

		msg, err := source.Recv()
		require.NoError(t, err)
		assert.Equal(t, "partial", msg.Content)

		_, err = source.Recv()
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestFromChannel(t *testing.T) {
	ch := make(chan Message, 1)
	ch <- Message{Role: RoleUser, Content: "test"}

	close(ch)

	builder := FromChannel(ch)
	assert.NotNil(t, builder)
	assert.NotNil(t, builder.source)
}

func TestFromCallback(t *testing.T) {
	builder := FromCallback(func(w CallbackWriter) error {
		w.WriteText("test")

		return nil
	})
	assert.NotNil(t, builder)
	assert.NotNil(t, builder.source)
}
