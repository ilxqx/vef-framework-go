// Package id provides various ID generation strategies for distributed systems.
// It supports multiple algorithms including Snowflake, XID, UUID v7, and custom random IDs.
//
// Quick start:
//
//	id := id.Generate()        // Uses XID (fastest)
//	uuid := id.GenerateUuid()  // Uses UUID v7 (standards compliant)
//
// For custom generators:
//
//	snowflake, _ := id.NewSnowflakeIdGenerator(1)  // For distributed systems
//	random := id.NewRandomIdGenerator("0-9a-z", 16) // Custom alphabet
package id

// IdGenerator defines the interface for all ID generation strategies.
// All generators must implement this interface to ensure consistency.
type IdGenerator interface {
	// Generate creates a new unique identifier as a string.
	// The format and characteristics depend on the specific implementation.
	Generate() string
}
