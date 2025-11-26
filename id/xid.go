package id

import "github.com/rs/xid"

// DefaultXidIdGenerator is the default XID generator instance.
var DefaultXidIdGenerator = NewXidIdGenerator()

type xidIdGenerator struct{}

// Generate creates a new XID as a 20-character base32-encoded string.
func (g *xidIdGenerator) Generate() string {
	return xid.New().String()
}

// NewXidIdGenerator creates a new XID generator instance.
func NewXidIdGenerator() IdGenerator {
	return &xidIdGenerator{}
}
