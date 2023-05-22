package client

import "runtime"

var (
	SelfPath   string
	Version    string
	SystemOS   = runtime.GOOS
	SystemUUID string
)
