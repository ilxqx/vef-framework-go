package id

// IdGenerator defines the interface for all ID generation strategies.
// All generators must implement this interface to ensure consistency.
type IdGenerator interface {
	// Generate creates a new unique identifier as a string.
	// The format and characteristics depend on the specific implementation.
	Generate() string
}
