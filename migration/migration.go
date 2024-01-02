package migration

import "embed"

//go:embed *.sql
var SQLFiles embed.FS
