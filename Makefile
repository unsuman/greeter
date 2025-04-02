all: clean build

build: build-main build-hindi build-japanese

build-main:
	go build -o greeter main.go

build-hindi:
	go build -o plugins/lang/hindi plugins/hindi/main.go

build-japanese:
	go build -o plugins/lang/japanese plugins/japanese/main.go

.PHONY: all build clean build-main build-hindi build-japanese