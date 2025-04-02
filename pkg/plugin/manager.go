package plugin

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
)

// PluginManager manages the lifecycle of plugins
type PluginManager struct {
	pluginsDir string
	plugins    map[string]*PluginInstance
	logger     *logrus.Logger
	mutex      sync.RWMutex
}

// PluginInstance represents a running plugin instance
type PluginInstance struct {
	Name       string
	Command    *exec.Cmd
	Stdin      io.WriteCloser
	Stdout     io.ReadCloser
	Client     *GRPCClient
	Logger     *logrus.Entry
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(logger *logrus.Logger, pluginsDir string) *PluginManager {
	return &PluginManager{
		pluginsDir: pluginsDir,
		plugins:    make(map[string]*PluginInstance),
		logger:     logger,
	}
}

// DiscoverPlugins finds all available plugins in the plugins directory
func (pm *PluginManager) DiscoverPlugins(category string) ([]string, error) {
	pm.logger.Infof("Discovering plugins in category: %s", category)

	pluginPath := filepath.Join(pm.pluginsDir, category)
	entries, err := os.ReadDir(pluginPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No plugins directory or empty
			return []string{}, nil
		}
		pm.logger.Errorf("Failed to read plugins directory: %v", err)
		return nil, fmt.Errorf("failed to read plugins directory: %w", err)
	}

	var plugins []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip directories
		}

		pluginName := entry.Name()
		execPath := filepath.Join(pluginPath, pluginName)

		// Check if the plugin binary exists and is executable
		if info, err := os.Stat(execPath); err == nil && !info.IsDir() {
			if info.Mode()&0111 != 0 { // Check if executable
				plugins = append(plugins, pluginName)
				pm.logger.Debugf("Found plugin: %s at %s", pluginName, execPath)
			}
		}
	}

	return plugins, nil
}

// StartPlugin launches a plugin process
func (pm *PluginManager) StartPlugin(category, name string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pluginKey := category + "-" + name

	if _, exists := pm.plugins[pluginKey]; exists {
		return nil // Plugin already running
	}

	execPath := filepath.Join(pm.pluginsDir, category, name)

	pm.logger.Infof("Starting plugin: %s (%s)", name, execPath)

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, execPath)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Capture stderr for logging
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("failed to start plugin: %w", err)
	}

	pluginLogger := pm.logger.WithField("plugin", name)

	// Log plugin stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			pluginLogger.Info(scanner.Text())
		}
	}()

	// Create gRPC client
	client, err := NewGRPCClient(stdin, stdout, pluginLogger)
	if err != nil {
		cancel()
		cmd.Process.Kill()
		return fmt.Errorf("failed to create gRPC client: %w", err)
	}

	instance := &PluginInstance{
		Name:       name,
		Command:    cmd,
		Stdin:      stdin,
		Stdout:     stdout,
		Client:     client,
		Logger:     pluginLogger,
		ctx:        ctx,
		cancelFunc: cancel,
	}

	pm.plugins[pluginKey] = instance

	// Handle process exit
	go func() {
		err := cmd.Wait()
		pm.mutex.Lock()
		delete(pm.plugins, pluginKey)
		pm.mutex.Unlock()

		if err != nil && ctx.Err() == nil { // Don't log if we cancelled the context
			pluginLogger.Errorf("Plugin exited with error: %v", err)
		} else {
			pluginLogger.Info("Plugin exited")
		}
	}()

	return nil
}

// StopPlugin terminates a plugin process
func (pm *PluginManager) StopPlugin(category, name string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pluginKey := category + "-" + name
	instance, exists := pm.plugins[pluginKey]
	if !exists {
		return nil // Plugin not running
	}

	pm.logger.Infof("Stopping plugin: %s", name)

	// Close gRPC client
	if instance.Client != nil {
		//first close pipes
		if err := instance.Stdin.Close(); err != nil {
			pm.logger.Warnf("Failed to close stdin pipe: %v", err)
		}
		if err := instance.Stdout.Close(); err != nil {
			pm.logger.Warnf("Failed to close stdout pipe: %v", err)
		}
		//then close client
		if err := instance.Client.Close(); err != nil {
			pm.logger.Warnf("Failed to close gRPC client: %v", err)
		}

	}

	pm.logger.Info("canceling plugin context")

	// Cancel the context to stop the command
	instance.cancelFunc()

	// Force kill if necessary
	pm.logger.Info("waiting for plugin to exit")
	if instance.Command.ProcessState == nil || !instance.Command.ProcessState.Exited() {
		if err := instance.Command.Process.Kill(); err != nil {
			pm.logger.Warnf("Failed to kill plugin process: %v", err)
		}
	}

	delete(pm.plugins, pluginKey)

	return nil
}

// GetGreeting sends a command to a plugin and returns the response
func (pm *PluginManager) GetGreeting(ctx context.Context, category, name, command string) (string, error) {
	pm.mutex.RLock()
	pluginKey := category + "-" + name
	instance, exists := pm.plugins[pluginKey]
	pm.mutex.RUnlock()

	if !exists {
		// Try to start the plugin if it's not running
		if err := pm.StartPlugin(category, name); err != nil {
			return "", fmt.Errorf("plugin %s is not running and could not be started: %w", pluginKey, err)
		}

		pm.mutex.RLock()
		instance = pm.plugins[pluginKey]
		pm.mutex.RUnlock()
	}

	pm.logger.Debugf("Executing command '%s' on plugin %s", command, name)

	// Use the gRPC client to get the greeting
	return instance.Client.GetGreeting(ctx, command)
}

// CleanupPlugins stops all running plugins
func (pm *PluginManager) CleanupPlugins() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	for key, instance := range pm.plugins {
		pm.logger.Infof("Stopping plugin: %s", instance.Name)

		// Close gRPC client
		if instance.Client != nil {
			//first close pipes
			if err := instance.Stdin.Close(); err != nil {
				pm.logger.Warnf("Failed to close stdin pipe: %v", err)
			}
			if err := instance.Stdout.Close(); err != nil {
				pm.logger.Warnf("Failed to close stdout pipe: %v", err)
			}
			//then close client
			if err := instance.Client.Close(); err != nil {
				pm.logger.Warnf("Failed to close gRPC client: %v", err)
			}

		}

		// Cancel the context to stop the command
		instance.cancelFunc()

		// Force kill if necessary
		if instance.Command.ProcessState == nil || !instance.Command.ProcessState.Exited() {
			if err := instance.Command.Process.Kill(); err != nil {
				pm.logger.Warnf("Failed to kill plugin process: %v", err)
			}
		}

		delete(pm.plugins, key)
	}
}
