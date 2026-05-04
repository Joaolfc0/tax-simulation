package tax

import (
	"encore.dev/storage/objects"
	"encore.dev/storage/sqldb"
)

var db = sqldb.NewDatabase("taxdb", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var uploadsBucket = objects.NewBucket("tax-uploads", objects.BucketConfig{})

var secrets struct {
	InfoSimplesToken string
}
