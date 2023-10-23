# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOOS = linux windows darwin
GODIST = dist

# Application name
APPNAME = modpack-manifest-downloader

# Version (optional)
VERSION = 1.0.0

# Targets
all: clean build

build: $(GOOS)

linux:
	@echo "Building for Linux"
	@mkdir -p $(GODIST)/linux
	$(GOBUILD) -o $(GODIST)/linux/$(APPNAME) main.go

windows:
	@echo "Building for Windows"
	@mkdir -p $(GODIST)/windows
	GOOS=windows $(GOBUILD) -o $(GODIST)/windows/$(APPNAME).exe main.go

darwin:
	@echo "Building for macOS"
	@mkdir -p $(GODIST)/darwin
	GOOS=darwin $(GOBUILD) -o $(GODIST)/darwin/$(APPNAME) main.go

clean:
	@echo "Cleaning up"
	@rm -rf $(GODIST)

.PHONY: clean