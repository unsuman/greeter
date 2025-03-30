package plugin

import (
	"context"
	"io"
)

// Plugin defines the interface that all plugins must implement
type Plugin interface {
	// Start initializes and starts the plugin
	Start(ctx context.Context) error

	// Stop gracefully shuts down the plugin
	Stop() error

	// Name returns the plugin's name
	Name() string
}

// GRPCPlugin defines additional methods for gRPC-based plugins
type GRPCPlugin interface {
	Plugin

	// Client creates a client from stdin/stdout
	Client(stdin io.WriteCloser, stdout io.ReadCloser) (interface{}, error)

	// Server returns a gRPC server
	Server(impl interface{}) (interface{}, error)
}
