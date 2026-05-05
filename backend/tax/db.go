package tax

import "encore.dev/storage/sqldb"

// db is the Encore-managed Postgres connection for the tax service.
// Migrations live in ./migrations (relative to this file's package).
var db = sqldb.NewDatabase("taxdb", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})
