package registry

import (
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/unsuman/greeter/pkg/greetings"
)

// Registry stores all available language plugins
type Registry struct {
	plugins map[string]greetings.Plugin
	logger  *logrus.Logger
	mu      sync.RWMutex
}

// New creates a new plugin registry
func New(logger *logrus.Logger) *Registry {
	return &Registry{
		plugins: make(map[string]greetings.Plugin),
		logger:  logger,
	}
}

// Register adds a plugin to the registry
func (r *Registry) Register(plugin greetings.Plugin) {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := plugin.Name()
	if _, exists := r.plugins[name]; exists {
		r.logger.Warnf("Plugin %s already registered, ignoring", name)
		return
	}

	if err := plugin.Init(); err != nil {
		r.logger.Errorf("Failed to initialize plugin %s: %v", name, err)
		return
	}

	r.plugins[name] = plugin
	r.logger.Infof("Registered plugin: %s", name)
}

// Get retrieves a plugin by name
func (r *Registry) Get(name string) (greetings.Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	plugin, exists := r.plugins[name]
	return plugin, exists
}

// List returns all registered plugin names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var names []string
	for name := range r.plugins {
		names = append(names, name)
	}
	return names
}

// Close closes all plugins and clears the registry
func (r *Registry) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for name, plugin := range r.plugins {
		if err := plugin.Close(); err != nil {
			r.logger.Warnf("Error closing plugin %s: %v", name, err)
		}
	}
	r.plugins = make(map[string]greetings.Plugin)
}

// The registry singleton for plugins to register with
var DefaultRegistry *Registry

// Initialize with a default logger
func init() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	DefaultRegistry = New(logger)
}

// Register is a convenience function for plugins to register with the default registry
func Register(plugin greetings.Plugin) {
	if DefaultRegistry != nil {
		DefaultRegistry.Register(plugin)
	}
}
