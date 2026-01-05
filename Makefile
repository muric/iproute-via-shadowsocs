# Variables
APP_NAME := srtunectl
SYSTEMD_DIR := /etc/systemd/system
SERVICE_NAME := route.service

IPROUTE_GIT_DIR := $(shell pwd)
# Default target
all: build

# Build target
build:
	@echo "Building the application..."
	go build -o ./output/${APP_NAME} main.go
# Clean target
clean:
	@echo "Cleaning up..."
	rm -rf ${APP_NAME}
	rm -rf /usr/bin/${APP_NAME}

install:
	install -m 0755 ./output/${APP_NAME} /usr/bin/$(APP_NAME)

	install -d $(SYSTEMD_DIR)
	sed \
		-e 's|@IPROUTE_GIT_DIR@|$(IPROUTE_GIT_DIR)|g' \
		route.service.in \
		> $(SYSTEMD_DIR)/$(SERVICE_NAME)

	systemctl daemon-reload
	systemctl enable $(SERVICE_NAME)

.PHONY: all build clean
