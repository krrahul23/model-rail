package constant

import "time"

// Default server settings used when the configuration omits them.
const (
	DefaultPublicAddr        = ":8080"
	DefaultAdminAddr         = ":8081"
	DefaultReadHeaderTimeout = 5 * time.Second
	DefaultShutdownGrace     = 15 * time.Second
)
