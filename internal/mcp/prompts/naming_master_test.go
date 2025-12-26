package prompts

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilxqx/vef-framework-go/mcp"
)

func TestNamingMasterPrompt(t *testing.T) {
	provider := NewNamingMasterPrompt()
	require.NotNil(t, provider)

	prompts := provider.Prompts()
	require.Len(t, prompts, 1, "Should have exactly one prompt definition")

	def := prompts[0]
	assert.NotNil(t, def.Prompt)
	assert.NotNil(t, def.Handler)

	assert.Equal(t, "naming-master", def.Prompt.Name)
	assert.Contains(t, def.Prompt.Description, "naming expert")
	assert.Contains(t, def.Prompt.Description, "database")

	ctx := context.Background()
	req := &mcp.GetPromptRequest{}

	result, err := def.Handler(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.NotEmpty(t, result.Description)
	assert.Len(t, result.Messages, 1, "Should have exactly one message")

	msg := result.Messages[0]
	assert.Equal(t, mcp.Role("user"), msg.Role)

	textContent, ok := msg.Content.(*mcp.TextContent)
	require.True(t, ok, "Message content should be TextContent")
	assert.NotEmpty(t, textContent.Text)

	content := textContent.Text
	assert.Contains(t, content, "Naming Master")
	assert.Contains(t, content, "Core Principles")
	assert.Contains(t, content, "Code Naming Conventions")
	assert.Contains(t, content, "Database Naming Conventions")
	assert.Contains(t, content, "Reserved Word")
	assert.Contains(t, content, "Interaction Protocol")
	assert.Contains(t, content, "Self-Check Checklist")
}

func TestNamingMasterPromptContent(t *testing.T) {
	assert.NotEmpty(t, namingMasterPromptContent, "Embedded naming-master.md content should not be empty")

	assert.Contains(t, namingMasterPromptContent, "# Naming Master")
	assert.Contains(t, namingMasterPromptContent, "## Style Matrix")
	assert.Contains(t, namingMasterPromptContent, "## Standard Audit Fields")
	assert.Contains(t, namingMasterPromptContent, "## Foreign Key Strategy Matrix")
	assert.Contains(t, namingMasterPromptContent, "## Index Design Considerations")
}
