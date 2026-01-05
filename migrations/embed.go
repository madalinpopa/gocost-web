package migrations

import "embed"

//go:embed *.sql
var MigrationFiles embed.FS
