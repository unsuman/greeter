# Default: build English-only version and all plugins
all: clean build-english build-plugins build-all

# Build with English only
build-english: build-plugins
	go build -o bin/greeter cmd/greeter/main.go

# Build with all languages
build-all:
	go build -o bin/greeter-all cmd/greeter-all/main.go

# Build external plugins
build-plugins: build-plugin-hindi build-plugin-japanese

# Build Hindi plugin
build-plugin-hindi:
	@mkdir -p bin/lang	
	go build -o bin/lang/hindi plugins/hindi/main.go

# Build Japanese plugin
build-plugin-japanese:
	@mkdir -p bin/lang
	go build -o bin/lang/japanese plugins/japanese/main.go

# Clean build artifacts
clean:
	rm -rf bin/

.PHONY: all clean build-english build-hindi build-japanese build-all build-plugins
