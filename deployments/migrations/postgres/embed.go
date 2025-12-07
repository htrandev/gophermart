package postgres

import "embed"

//go:embed *.sql
var Embed embed.FS
