package id

// IDGenerator defines the interface for all ID generation strategies.
// All generators must implement this interface to ensure consistency.
type IDGenerator interface {
	// Generate creates a new unique identifier as a string.
	// The format and characteristics depend on the specific implementation.
	Generate() string
}
