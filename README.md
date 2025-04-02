# Greeter - A Multilingual Greeting Application

Greeter is a command-line application that provides greetings in multiple languages through a plugin-based architecture. It includes a built-in English language implementation and supports additional languages via external plugins.

![](assets/example.gif)

## Overview

The Greeter application allows users to get greetings like "hello", "good morning", etc. in different languages. It has a modular architecture with:

- A core application that handles user input and manages plugins
- A built-in English language module
- Support for additional languages via plugins (currently includes Hindi & Japanese)
- A plugin system based on gRPC for inter-process communication

## Architecture
<img alt="Plugin" src="/assets/plugin.png">

- The main application (greeter) processes user commands and routes them to either the built-in languages or plugins
- The Plugin Manager handles discovery, launching, and communication with language plugins
- Plugins are separate executables that implement the GreeterService gRPC interface
- Communication between the main app and plugins is done via stdin/stdout pipes and gRPC

## Plugin Communication

Plugins communicate with the main application using gRPC over stdin/stdout. The protocol is defined in greeter.proto.

When a plugin starts, it:
1. Initializes the gRPC server
2. Writes its process ID to stdout for the main app to track
3. Listens for incoming gRPC requests
4. Processes greeting requests and returns appropriate responses

## Building the Project

### Prerequisites

- Go 1.23 or later
- Make (optional, for using the Makefile)

### Build Instructions

1. Clone the repository:
   ```bash
   git clone https://github.com/unsuman/greeter.git
   cd greeter
   ```

2. Build the project:
   ```bash
   # Using Make
   make

   # Or using Go directly
   go build -o greeter main.go
   go build -o plugins/lang/hindi plugins/hindi/main.go
   ```

## Installation

After building, you can place the binaries in some directory or keep them in a development directory:

```
./greeter                     # Main executable
./plugins/lang/hindi          # Hindi plugin
```

## Configuration

### Environment Variables
Set `GREETER_PLUGIN_PATH` to the plugins binary directory.

## Usage

### Basic Commands

```bash
# Get a greeting in English (default)
greeter hello

# Get a greeting in Hindi
greeter hello --lang=hindi

# Get a greeting in Japanese
greeter hello --lang=japanese

# List available languages
greeter list-languages

# Shutdown a running plugin
greeter shutdown-plugin --lang=hindi

# Get other greetings
greeter goodmorning [--lang=language]
greeter goodafternoon [--lang=language]
greeter goodnight [--lang=language]
greeter goodbye [--lang=language]
```

## Plugin Management

The greeter application intelligently manages plugins:

1. **Plugin Persistence**: When you use a language plugin, it remains running in the background for future use.
2. **Plugin Reuse**: If a plugin is already running, the application connects to the existing process instead of starting a new one.
3. **Explicit Shutdown**: You can shut down a running plugin with the `shutdown-plugin` command.

This approach provides several benefits:
- Faster response times for repeated commands
- Lower resource usage
- Ability to manage plugin lifecycle when needed

## Current Limitations

1. **No Cross-User Plugin Sharing**: Plugins started by one user can't be used by another.
2. **Limited Plugin Discovery**: The system can only detect plugins started by the same application instance.
3. **Basic Error Handling**: Error handling is minimal, especially for plugin communication failures.
4. **Limited Configuration**: Plugin parameters cannot be configured without modifying the source code.
5. **No Automatic Plugin Cleanup**: Plugins must be explicitly shut down, or they'll remain running until the system is restarted.

## Adding New Language Plugins

To create a new language plugin:

1. Create a new directory in plugins (e.g., `plugins/french/`)
2. Implement the GreeterService interface (see main.go as an example)
3. Build the plugin and place it in the appropriate location (`plugins/lang/french`)

The minimal implementation needs:
- A gRPC server that implements the GreeterService
- Functions for each greeting type (Hello, GoodMorning, etc.)
- Proper handling of stdin/stdout for communication with the main app


