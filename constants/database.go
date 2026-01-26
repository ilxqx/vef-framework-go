package constants

// DBType represents supported database types.
type DBType string

// Supported database types.
const (
	Oracle    DBType = "oracle"
	SQLServer DBType = "sqlserver"
	Postgres  DBType = "postgres"
	MySQL     DBType = "mysql"
	SQLite    DBType = "sqlite"
)
