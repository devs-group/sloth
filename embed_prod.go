//go:build production

package main

import "embed"

//go:embed all:frontend/.output/public/*
var VueFiles embed.FS
