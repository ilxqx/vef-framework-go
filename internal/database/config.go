package database

import (
	"runtime"
	"time"
)

// ConnectionPoolConfig defines the connection pool configuration.
type ConnectionPoolConfig struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration
}

// Default connection pool constants.
const (
	DefaultMaxIdleConnsMultiplier = 4
	DefaultMaxOpenConnsMultiplier = 16
	DefaultMinIdleConns           = 25
	DefaultMinOpenConns           = 100
	DefaultConnMaxIdleTime        = 5 * time.Minute
	DefaultConnMaxLifetime        = 30 * time.Minute
)

// NewDefaultConnectionPoolConfig creates a default connection pool configuration.
func NewDefaultConnectionPoolConfig() *ConnectionPoolConfig {
	return &ConnectionPoolConfig{
		MaxIdleConns:    max(runtime.GOMAXPROCS(0)*DefaultMaxIdleConnsMultiplier, DefaultMinIdleConns),
		MaxOpenConns:    max(runtime.GOMAXPROCS(0)*DefaultMaxOpenConnsMultiplier, DefaultMinOpenConns),
		ConnMaxIdleTime: DefaultConnMaxIdleTime,
		ConnMaxLifetime: DefaultConnMaxLifetime,
	}
}

// ApplyToDB applies the connection pool configuration to a database connection.
func (c *ConnectionPoolConfig) ApplyToDB(db interface {
	SetMaxIdleConns(int)
	SetMaxOpenConns(int)
	SetConnMaxIdleTime(time.Duration)
	SetConnMaxLifetime(time.Duration)
},
) {
	db.SetMaxIdleConns(c.MaxIdleConns)
	db.SetMaxOpenConns(c.MaxOpenConns)
	db.SetConnMaxIdleTime(c.ConnMaxIdleTime)
	db.SetConnMaxLifetime(c.ConnMaxLifetime)
}
