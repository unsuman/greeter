package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"

	// Import only English as embedded

	"github.com/unsuman/greeter/pkg/cmd"
	"github.com/unsuman/greeter/pkg/plugin"
	_ "github.com/unsuman/greeter/plugins/english/pkg"

	"github.com/unsuman/greeter/pkg/plugin/registry"
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

	// Setup plugins
	pluginsDir, pluginMgr = cmd.SetupPlugins(log)
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Info("Shutting down...")
		registry.DefaultRegistry.Close()
		pluginMgr.CleanupPlugins()
		os.Exit(0)
	}()

	if len(os.Args) < 2 {
		cmd.PrintUsage()
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
		cmd.ListAvailableLanguages(log, pluginMgr)
		return
	}

	// Get greeting
	message, err := cmd.GetGreeting(log, pluginMgr, pluginsDir, command, language)
	if err != nil {
		log.Errorf("Failed to get greeting: %v", err)
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	green := "\033[92m"
	reset := "\033[0m"

	fmt.Println(green + message + reset)

	pluginMgr.StopPlugin("lang", language)
	registry.DefaultRegistry.Close()

	log.Info("Exiting...")
}
