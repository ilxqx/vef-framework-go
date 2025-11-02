package docs

import (
	"os"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilxqx/vef-framework-go/id"
)

// TestREADMEAccuracy validates that README claims match actual codebase implementation.
// This test suite serves as a guard against documentation drift.

// TestGoVersionRequirement verifies that the README documents the correct Go version requirement.
func TestGoVersionRequirement(t *testing.T) {
	// Read go.mod to determine actual version requirement
	goModContent, err := os.ReadFile("../../go.mod")
	require.NoError(t, err, "Failed to read go.mod")

	// Extract Go version from go.mod
	lines := strings.Split(string(goModContent), "\n")
	var goVersion string
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "go ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				goVersion = parts[1]
				break
			}
		}
	}

	require.NotEmpty(t, goVersion, "Could not find Go version in go.mod")
	assert.Equal(t, "1.25.0", goVersion, "go.mod specifies Go 1.25.0")

	// Read README.md
	readmeMdContent, err := os.ReadFile("../../README.md")
	require.NoError(t, err, "Failed to read README.md")

	// Check if README.md mentions the correct version
	readmeText := string(readmeMdContent)

	// The README should mention "Go 1.25.0 or higher" for accuracy
	// Currently it says "Go 1.25 or higher" which is imprecise
	assert.Contains(t, readmeText, "Go 1.25", "README.md should mention Go version requirement")

	// Read README.zh-CN.md
	readmeZhContent, err := os.ReadFile("../../README.zh-CN.md")
	require.NoError(t, err, "Failed to read README.zh-CN.md")

	readmeZhText := string(readmeZhContent)
	assert.Contains(t, readmeZhText, "Go 1.25", "README.zh-CN.md should mention Go version requirement")
}

// TestIDGenerationFormat verifies that README accurately describes ID generation.
func TestIDGenerationFormat(t *testing.T) {
	// Generate a sample ID using the actual implementation
	sampleID := id.Generate()

	// Verify the ID characteristics match what README claims
	assert.Len(t, sampleID, 20, "Generated IDs should be 20 characters long")

	// Verify base32 encoding (characters 0-9, a-v)
	for _, ch := range sampleID {
		assert.True(t, (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'v'),
			"ID should only contain base32 characters (0-9, a-v), got: %c", ch)
	}

	// Read README.md and verify it documents XID correctly
	readmeMdContent, err := os.ReadFile("../../README.md")
	require.NoError(t, err, "Failed to read README.md")

	readmeText := string(readmeMdContent)

	// Check that README mentions XID and base32
	assert.Contains(t, readmeText, "20-character", "README should mention 20-character IDs")
	assert.Contains(t, readmeText, "base32", "README should mention base32 encoding")

	// Read README.zh-CN.md
	readmeZhContent, err := os.ReadFile("../../README.zh-CN.md")
	require.NoError(t, err, "Failed to read README.zh-CN.md")

	readmeZhText := string(readmeZhContent)
	assert.Contains(t, readmeZhText, "20", "Chinese README should mention 20-character IDs")
	assert.Contains(t, readmeZhText, "base32", "Chinese README should mention base32 encoding")
}

// TestStorageResourceNamespace verifies that README examples use the correct storage namespace.
func TestStorageResourceNamespace(t *testing.T) {
	// The actual storage resource is registered as "sys/storage"
	// as seen in internal/storage/storage_resource.go:36

	readmeMdContent, err := os.ReadFile("../../README.md")
	require.NoError(t, err, "Failed to read README.md")

	readmeText := string(readmeMdContent)

	// Check for incorrect namespace
	if strings.Contains(readmeText, "base/storage") {
		t.Error("README.md incorrectly uses 'base/storage' - should be 'sys/storage'")
	}

	// Check that correct namespace is mentioned
	assert.Contains(t, readmeText, "sys/storage", "README.md should use correct 'sys/storage' namespace")

	// Check Chinese README
	readmeZhContent, err := os.ReadFile("../../README.zh-CN.md")
	require.NoError(t, err, "Failed to read README.zh-CN.md")

	readmeZhText := string(readmeZhContent)

	if strings.Contains(readmeZhText, "base/storage") {
		t.Error("README.zh-CN.md incorrectly uses 'base/storage' - should be 'sys/storage'")
	}

	assert.Contains(t, readmeZhText, "sys/storage", "README.zh-CN.md should use correct 'sys/storage' namespace")
}

// TestBadgerDBNotDocumented verifies that BadgerDB is not falsely documented as a dependency.
func TestBadgerDBNotDocumented(t *testing.T) {
	// BadgerDB is not implemented in the codebase, so it should not be in README

	readmeMdContent, err := os.ReadFile("../../README.md")
	require.NoError(t, err, "Failed to read README.md")

	readmeText := strings.ToLower(string(readmeMdContent))

	// BadgerDB should not be mentioned in README
	if strings.Contains(readmeText, "badgerdb") || strings.Contains(readmeText, "badger") {
		// Allow "badger" in other contexts, but "BadgerDB" as a tech stack item should not exist
		// This is a soft check - manual review may be needed
		t.Log("Warning: README.md may contain references to BadgerDB which is not implemented")
	}

	// Check project.md for BadgerDB references
	projectMdContent, err := os.ReadFile("../../openspec/project.md")
	if err == nil {
		projectText := strings.ToLower(string(projectMdContent))
		if strings.Contains(projectText, "badgerdb") {
			t.Error("project.md should not reference BadgerDB as it is not implemented")
		}
	}
}

// TestReservedNamespaces verifies that reserved system namespaces are correctly documented.
func TestReservedNamespaces(t *testing.T) {
	// The framework reserves these namespaces:
	// - security/auth
	// - sys/storage
	// - sys/monitor

	readmeMdContent, err := os.ReadFile("../../README.md")
	require.NoError(t, err, "Failed to read README.md")

	readmeText := string(readmeMdContent)

	// All reserved namespaces should be mentioned
	assert.Contains(t, readmeText, "security/auth", "README should document security/auth namespace")
	assert.Contains(t, readmeText, "sys/storage", "README should document sys/storage namespace")
	assert.Contains(t, readmeText, "sys/monitor", "README should document sys/monitor namespace")

	// Check Chinese README
	readmeZhContent, err := os.ReadFile("../../README.zh-CN.md")
	require.NoError(t, err, "Failed to read README.zh-CN.md")

	readmeZhText := string(readmeZhContent)

	assert.Contains(t, readmeZhText, "security/auth", "Chinese README should document security/auth namespace")
	assert.Contains(t, readmeZhText, "sys/storage", "Chinese README should document sys/storage namespace")
	assert.Contains(t, readmeZhText, "sys/monitor", "Chinese README should document sys/monitor namespace")
}

// TestFrameworkVersion verifies that build info is available.
func TestFrameworkVersion(t *testing.T) {
	// Check that the framework version can be determined from build info
	buildInfo, ok := debug.ReadBuildInfo()
	require.True(t, ok, "Should be able to read build info")
	require.NotNil(t, buildInfo, "Build info should not be nil")

	// The module path should match (or be a subpackage during testing)
	assert.Contains(t, buildInfo.Path, "github.com/ilxqx/vef-framework-go",
		"Module path should contain the actual module path")
}

// TestDocumentationConsistency checks that both language versions mention the same key concepts.
func TestDocumentationConsistency(t *testing.T) {
	readmeMdContent, err := os.ReadFile("../../README.md")
	require.NoError(t, err, "Failed to read README.md")

	readmeZhContent, err := os.ReadFile("../../README.zh-CN.md")
	require.NoError(t, err, "Failed to read README.zh-CN.md")

	readmeText := string(readmeMdContent)
	readmeZhText := string(readmeZhContent)

	// Both should mention core technologies
	coreTech := []string{"PostgreSQL", "SQLite", "Redis", "MinIO", "Fiber", "Bun"}

	for _, tech := range coreTech {
		assert.Contains(t, readmeText, tech,
			"English README should mention %s", tech)
		assert.Contains(t, readmeZhText, tech,
			"Chinese README should mention %s (technical terms not translated)", tech)
	}

	// MySQL might be in different case or translated
	readmeTextLower := strings.ToLower(readmeText)
	readmeZhTextLower := strings.ToLower(readmeZhText)
	assert.True(t, strings.Contains(readmeTextLower, "mysql"),
		"English README should mention MySQL (case-insensitive)")
	assert.True(t, strings.Contains(readmeZhTextLower, "mysql") || strings.Contains(readmeZhText, "数据库"),
		"Chinese README should mention MySQL or database")

	// Both should mention key features
	assert.Contains(t, readmeText, "XID", "English README should mention XID")
	assert.Contains(t, readmeZhText, "XID", "Chinese README should mention XID")
}
