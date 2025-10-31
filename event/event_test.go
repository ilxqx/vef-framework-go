package event

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBaseEvent(t *testing.T) {
	t.Run("MinimalCreationWithJustType", func(t *testing.T) {
		event := NewBaseEvent("test.event")

		assert.Equal(t, "test.event", event.Type())
		assert.Equal(t, "", event.Source())
		assert.NotEmpty(t, event.Id())
		assert.False(t, event.Time().IsZero())
		assert.Equal(t, 0, len(event.Meta()))
	})

	t.Run("CreationWithSourceOption", func(t *testing.T) {
		event := NewBaseEvent("user.created", WithSource("user-service"))

		assert.Equal(t, "user.created", event.Type())
		assert.Equal(t, "user-service", event.Source())
		assert.NotEmpty(t, event.Id())
		assert.False(t, event.Time().IsZero())
		assert.Equal(t, 0, len(event.Meta()))
	})

	t.Run("CreationWithSingleMetadataOption", func(t *testing.T) {
		event := NewBaseEvent("order.placed", WithMeta("version", "1.0"))

		assert.Equal(t, "order.placed", event.Type())
		assert.Equal(t, "", event.Source())
		assert.NotEmpty(t, event.Id())
		assert.False(t, event.Time().IsZero())
		assert.Equal(t, 1, len(event.Meta()))
		assert.Equal(t, "1.0", event.Meta()["version"])
	})

	t.Run("CreationWithMultipleOptions", func(t *testing.T) {
		event := NewBaseEvent("payment.processed",
			WithSource("payment-service"),
			WithMeta("amount", "100.00"),
			WithMeta("currency", "USD"),
			WithMeta("gateway", "stripe"),
		)

		assert.Equal(t, "payment.processed", event.Type())
		assert.Equal(t, "payment-service", event.Source())
		assert.NotEmpty(t, event.Id())
		assert.False(t, event.Time().IsZero())

		meta := event.Meta()
		assert.Equal(t, 3, len(meta))
		assert.Equal(t, "100.00", meta["amount"])
		assert.Equal(t, "USD", meta["currency"])
		assert.Equal(t, "stripe", meta["gateway"])
	})

	t.Run("EachEventHasUniqueIDAndTime", func(t *testing.T) {
		event1 := NewBaseEvent("test.event")

		time.Sleep(1 * time.Millisecond)

		event2 := NewBaseEvent("test.event")

		assert.NotEqual(t, event1.Id(), event2.Id())
		assert.True(t, event2.Time().After(event1.Time()) || event2.Time().Equal(event1.Time()))
	})
}

func TestBaseEvent_Metadata(t *testing.T) {
	t.Run("MetaReturnsAllMetadata", func(t *testing.T) {
		event := NewBaseEvent("test.event",
			WithMeta("key1", "value1"),
			WithMeta("key2", "value2"),
			WithMeta("key3", "value3"),
		)

		meta := event.Meta()
		assert.Equal(t, 3, len(meta))
		assert.Equal(t, "value1", meta["key1"])
		assert.Equal(t, "value2", meta["key2"])
		assert.Equal(t, "value3", meta["key3"])
	})

	t.Run("MetaReturnsCopyToPreventExternalModification", func(t *testing.T) {
		event := NewBaseEvent("test.event", WithMeta("key", "value"))

		meta := event.Meta()
		meta["key"] = "modified"
		meta["new"] = "added"

		originalMeta := event.Meta()
		assert.Equal(t, "value", originalMeta["key"])
		assert.NotContains(t, originalMeta, "new")
	})
}

func TestBaseEvent_JSONSerialization(t *testing.T) {
	t.Run("MarshalMinimalEvent", func(t *testing.T) {
		event := NewBaseEvent("test.event")

		jsonData, err := json.Marshal(event)
		require.NoError(t, err)

		var jsonMap map[string]any

		err = json.Unmarshal(jsonData, &jsonMap)
		require.NoError(t, err)

		assert.Equal(t, "test.event", jsonMap["type"])
		assert.Equal(t, "", jsonMap["source"])
		assert.NotEmpty(t, jsonMap["id"])
		assert.NotEmpty(t, jsonMap["time"])

		_, hasMetadata := jsonMap["metadata"]
		assert.False(t, hasMetadata)
	})

	t.Run("MarshalEventWithAllFields", func(t *testing.T) {
		event := NewBaseEvent("user.registered",
			WithSource("user-service"),
			WithMeta("version", "1.0"),
			WithMeta("region", "us-east-1"),
		)

		jsonData, err := json.Marshal(event)
		require.NoError(t, err)

		var jsonMap map[string]any

		err = json.Unmarshal(jsonData, &jsonMap)
		require.NoError(t, err)

		assert.Equal(t, "user.registered", jsonMap["type"])
		assert.Equal(t, "user-service", jsonMap["source"])
		assert.NotEmpty(t, jsonMap["id"])
		assert.NotEmpty(t, jsonMap["time"])

		metadata, ok := jsonMap["metadata"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "1.0", metadata["version"])
		assert.Equal(t, "us-east-1", metadata["region"])
	})

	t.Run("UnmarshalMinimalEvent", func(t *testing.T) {
		jsonData := `{
			"type": "test.unmarshal",
			"id": "test-id-123",
			"source": "test-source",
			"time": "2023-01-01T12:00:00Z"
		}`

		var event BaseEvent

		err := json.Unmarshal([]byte(jsonData), &event)
		require.NoError(t, err)

		assert.Equal(t, "test.unmarshal", event.Type())
		assert.Equal(t, "test-id-123", event.Id())
		assert.Equal(t, "test-source", event.Source())

		expectedTime, _ := time.Parse(time.RFC3339, "2023-01-01T12:00:00Z")
		assert.Equal(t, expectedTime, event.Time())
		assert.Equal(t, 0, len(event.Meta()))
	})

	t.Run("UnmarshalEventWithMetadata", func(t *testing.T) {
		jsonData := `{
			"type": "order.created",
			"id": "order-456",
			"source": "order-service",
			"time": "2023-06-15T10:30:00Z",
			"metadata": {
				"customer_id": "123",
				"total": "99.99"
			}
		}`

		var event BaseEvent

		err := json.Unmarshal([]byte(jsonData), &event)
		require.NoError(t, err)

		assert.Equal(t, "order.created", event.Type())
		assert.Equal(t, "order-456", event.Id())
		assert.Equal(t, "order-service", event.Source())

		meta := event.Meta()
		assert.Equal(t, 2, len(meta))
		assert.Equal(t, "123", meta["customer_id"])
		assert.Equal(t, "99.99", meta["total"])
	})

	t.Run("RoundtripSerializationPreservesData", func(t *testing.T) {
		original := NewBaseEvent("roundtrip.test",
			WithSource("test-service"),
			WithMeta("key1", "value1"),
			WithMeta("key2", "value2"),
		)

		jsonData, err := json.Marshal(original)
		require.NoError(t, err)

		var restored BaseEvent

		err = json.Unmarshal(jsonData, &restored)
		require.NoError(t, err)

		assert.Equal(t, original.Type(), restored.Type())
		assert.Equal(t, original.Id(), restored.Id())
		assert.Equal(t, original.Source(), restored.Source())
		assert.Equal(t, original.Time().Unix(), restored.Time().Unix())
		assert.Equal(t, original.Meta(), restored.Meta())
	})

	t.Run("UnmarshalHandlesMissingMetadataGracefully", func(t *testing.T) {
		jsonData := `{
			"type": "simple.event",
			"id": "simple-123",
			"source": "simple-service",
			"time": "2023-01-01T00:00:00Z"
		}`

		var event BaseEvent

		err := json.Unmarshal([]byte(jsonData), &event)
		require.NoError(t, err)

		assert.NotNil(t, event.Meta())
		assert.Equal(t, 0, len(event.Meta()))

		assert.Equal(t, 0, len(event.Meta()))
	})

	t.Run("UnmarshalInvalidJSONReturnsError", func(t *testing.T) {
		invalidJSON := `{invalid json`

		var event BaseEvent

		err := json.Unmarshal([]byte(invalidJSON), &event)
		assert.Error(t, err)
	})
}

func TestBaseEvent_Immutability(t *testing.T) {
	t.Run("CoreFieldsAreImmutableAfterCreation", func(t *testing.T) {
		event := NewBaseEvent("test.event", WithSource("test-source"))

		originalType := event.Type()
		originalId := event.Id()
		originalSource := event.Source()
		originalTime := event.Time()

		time.Sleep(1 * time.Millisecond)

		assert.Equal(t, originalType, event.Type())
		assert.Equal(t, originalId, event.Id())
		assert.Equal(t, originalSource, event.Source())
		assert.Equal(t, originalTime, event.Time())
	})

	t.Run("MetadataIsImmutableAfterCreation", func(t *testing.T) {
		event := NewBaseEvent("test.event", WithMeta("initial", "value"))

		meta := event.Meta()
		assert.Equal(t, 1, len(meta))
		assert.Equal(t, "value", meta["initial"])
	})
}
