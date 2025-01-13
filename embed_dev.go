//go:build !production

package main

import "embed"

// During development we don't embed any files
var VueFiles embed.FS
