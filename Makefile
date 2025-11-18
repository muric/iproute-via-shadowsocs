# Variables
APP_NAME = srtunectl

# Default target
all: build

# Build target
build:
	@echo "Building the application..."
	go build
# Clean target
clean:
	@echo "Cleaning up..."
	rm -rf ${APP_NAME}
	rm -rf /usr/bin/${APP_NAME}

install:
	cp ${APP_NAME} /usr/bin/

.PHONY: all build clean
