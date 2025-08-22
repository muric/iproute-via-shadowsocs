# Variables
APP_NAME = iproute

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

# Run target
run: build
        @echo "Running the application..."
        ./$(BUILD_TARGET)
install:
        cp ${APP_NAME} /usr/bin/

.PHONY: all build clean run install
