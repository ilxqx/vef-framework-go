package id

import "github.com/rs/xid"

// DefaultXidIdGenerator is the default XID generator instance.
// XID generates globally unique IDs with good performance characteristics.
// It produces 20-character strings using base32 encoding (0-9, a-v).
var DefaultXidIdGenerator = NewXidIdGenerator()

// xidIdGenerator implements IdGenerator using the XID algorithm.
// XID creates sortable unique identifiers that are URL-safe and compact.
type xidIdGenerator struct{}

// Generate creates a new XID as a 20-character base32-encoded string.
// XID combines timestamp, machine ID, process ID, and counter for uniqueness.
// The result is lexicographically sortable by creation time.
func (g *xidIdGenerator) Generate() string {
	return xid.New().String()
}

// NewXidIdGenerator creates a new XID generator instance.
// XID is recommended for high-performance scenarios where you need:
//   - Fast generation (best performance among all generators)
//   - Compact representation (20 characters)
//   - Time-based sorting
//   - No coordination between nodes
//
// Example:
//
//	gen := NewXidIdGenerator()
//	id := gen.Generate()  // Returns something like "9m4e2mr0ui3e8a215n4g"
func NewXidIdGenerator() IdGenerator {
	return &xidIdGenerator{}
}
