package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/unsuman/greeter/pkg/greetings"
	"github.com/unsuman/greeter/pkg/lang"
	"github.com/unsuman/greeter/pkg/plugin"
)

var (
	log        *logrus.Logger
	pluginMgr  *plugin.PluginManager
	pluginsDir string
)

func init() {
	log = logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Set log level based on environment variable
	logLevel := os.Getenv("GREETER_LOG_LEVEL")
	if logLevel == "debug" {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	// Determine plugins directory
	// Look for plugins in ${PREFIX}/libexec/greeter/plugins/lang/${PLUGIN_NAME}
	// For development, we'll use a directory relative to the binary
	execPath, err := os.Executable()
	if err != nil {
		log.Fatal("Failed to get executable path:", err)
	}

	// Check if GREETER_PLUGIN_PATH env var is set
	envPluginsDir := os.Getenv("GREETER_PLUGIN_PATH")
	if envPluginsDir != "" {
		pluginsDir = envPluginsDir
	} else {
		// Default to a directory relative to the binary
		pluginsDir = filepath.Join(filepath.Dir(execPath), "plugins")
	}

	log.Debugf("Using plugins directory: %s", pluginsDir)

	// Initialize plugin manager
	pluginMgr = plugin.NewPluginManager(log, pluginsDir)
}

func main() {
	// Set up signal handling for cleanup
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Info("Shutting down...")
		pluginMgr.CleanupPlugins()
		os.Exit(1)
	}()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := strings.ToLower(os.Args[1])
	language := "english" // Default language

	// Check if language is specified
	if len(os.Args) > 2 && strings.HasPrefix(os.Args[2], "--lang=") {
		language = strings.TrimPrefix(os.Args[2], "--lang=")
	}

	// Process commands
	switch command {
	case "list-languages":
		listAvailableLanguages()
		return
	}

	// Get greeting
	var message string
	var err error

	if language == "english" {
		// Use the built-in English implementation
		greeter := lang.NewEnglish()
		message = getGreetingFromGreeter(command, greeter)
	} else {
		// Use language plugin
		message, err = getGreetingFromPlugin(command, language)
		if err != nil {
			log.Errorf("Failed to get greeting from plugin: %v", err)
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println(message)

	pluginMgr.CleanupPlugins()
	log.Info("Exiting...")
}

func getGreetingFromGreeter(command string, greeter greetings.Greeter) string {
	switch command {
	case "hello":
		return greeter.Hello()
	case "goodmorning":
		return greeter.GoodMorning()
	case "goodafternoon":
		return greeter.GoodAfternoon()
	case "goodnight":
		return greeter.GoodNight()
	case "goodbye":
		return greeter.GoodBye()
	default:
		fmt.Println("Unknown greeting command:", command)
		printUsage()
		os.Exit(1)
		return ""
	}
}

func getGreetingFromPlugin(command, language string) (string, error) {
	// Check if plugin exists
	pluginPath := filepath.Join(pluginsDir, "lang", language)
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return "", fmt.Errorf("language plugin '%s' not found", language)
	}

	log.Debugf("Found plugin for language: %s at %s", language, pluginPath)

	// Try to start plugin if not already running
	if err := pluginMgr.StartPlugin("lang", language); err != nil {
		return "", fmt.Errorf("failed to start %s plugin: %w", language, err)
	}

	// Execute greeting command via gRPC
	ctx := context.Background()
	return pluginMgr.GetGreeting(ctx, "lang", language, command)
}

func listAvailableLanguages() {
	fmt.Println("Available languages:")
	fmt.Println("- english (built-in)")

	languages, err := pluginMgr.DiscoverPlugins("lang")
	if err != nil {
		log.Errorf("Failed to discover language plugins: %v", err)
		return
	}

	for _, lang := range languages {
		fmt.Printf("- %s (plugin)\n", lang)
	}
}

func printUsage() {
	fmt.Println("Usage: greeter <command> [--lang=language]")
	fmt.Println("Available commands: hello, goodmorning, goodafternoon, goodnight, goodbye, list-languages")
	fmt.Println("Example: greeter hello --lang=hindi")
}
