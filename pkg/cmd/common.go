package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/unsuman/greeter/pkg/greetings"
	"github.com/unsuman/greeter/pkg/plugin"
	"github.com/unsuman/greeter/pkg/plugin/registry"
)

// SetupPlugins initializes the plugin system
func SetupPlugins(logger *logrus.Logger) (string, *plugin.PluginManager) {
	execPath, err := os.Executable()
	if err != nil {
		logger.Fatal("Failed to get executable path:", err)
	}

	var pluginsDir string
	pluginsDir = filepath.Dir(execPath)

	logger.Infof("Using plugins directory: %s", pluginsDir)

	pluginMgr := plugin.NewPluginManager(logger, pluginsDir)

	return pluginsDir, pluginMgr
}

// GetGreeting gets a greeting from either an internal or external plugin
func GetGreeting(logger *logrus.Logger, pluginMgr *plugin.PluginManager, pluginsDir, command, language string) (string, error) {
	if plugin, exists := registry.DefaultRegistry.Get(language); exists {
		logger.Debugf("Using embedded plugin for language: %s", language)
		return GetGreetingFromInternalPlugin(command, plugin)
	}

	// If not found as embedded, try external plugin
	return GetGreetingFromExternalPlugin(logger, pluginMgr, pluginsDir, command, language)
}

// GetGreetingFromInternalPlugin gets a greeting from an internal plugin
func GetGreetingFromInternalPlugin(command string, plugin greetings.Plugin) (string, error) {
	switch command {
	case "hello":
		return plugin.Hello(), nil
	case "goodmorning":
		return plugin.GoodMorning(), nil
	case "goodafternoon":
		return plugin.GoodAfternoon(), nil
	case "goodnight":
		return plugin.GoodNight(), nil
	case "goodbye":
		return plugin.GoodBye(), nil
	default:
		return "", fmt.Errorf("unknown greeting command: %s", command)
	}
}

// GetGreetingFromExternalPlugin gets a greeting from an external plugin
func GetGreetingFromExternalPlugin(logger *logrus.Logger, pluginMgr *plugin.PluginManager, pluginsDir, command, language string) (string, error) {
	pluginPath := filepath.Join(pluginsDir, "lang", language)
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return "", fmt.Errorf("language plugin '%s' not found", language)
	}

	logger.Debugf("Found external plugin for language: %s at %s", language, pluginPath)

	if err := pluginMgr.StartPlugin("lang", language); err != nil {
		return "", fmt.Errorf("failed to start %s plugin: %w", language, err)
	}

	// Execute greeting command via gRPC
	ctx := context.Background()
	result, err := pluginMgr.GetGreeting(ctx, "lang", language, command)

	if err != nil {
		pluginMgr.StopPlugin("lang", language)
		return "", err
	}

	return result, nil
}

// ListAvailableLanguages lists all available languages
func ListAvailableLanguages(logger *logrus.Logger, pluginMgr *plugin.PluginManager) {
	fmt.Println("Available languages:")

	// List embedded languages first
	for _, lang := range registry.DefaultRegistry.List() {
		fmt.Printf("- %s (built-in)\n", lang)
	}

	// List external languages
	languages, err := pluginMgr.DiscoverPlugins("lang")
	if err != nil {
		logger.Errorf("Failed to discover language plugins: %v", err)
		return
	}

	// Don't show languages that are already listed as built-in
	for _, lang := range languages {
		if _, exists := registry.DefaultRegistry.Get(lang); !exists {
			fmt.Printf("- %s (plugin)\n", lang)
		}
	}
}

func PrintUsage() {
	fmt.Println("Usage: greeter <command> [--lang=language]")
	fmt.Println("Available commands: hello, goodmorning, goodafternoon, goodnight, goodbye, list-languages, shutdown-plugin")
	fmt.Println("Example: greeter hello --lang=hindi")
}
