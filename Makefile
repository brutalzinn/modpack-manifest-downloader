# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOOS = linux windows darwin
GOARCH = amd64 arm64
GODIST = dist
ZIP = zip

# Application name
APPNAME = modpack-manifest-downloader

# Version (optional)
VERSION = 0.0.1

# Targets
all: clean build zip

build: $(GOOS)

linux:
	@echo "Building for Linux"
	@mkdir -p $(GODIST)/linux
	$(foreach arch, $(GOARCH), $(GOBUILD) -o $(GODIST)/linux/$(APPNAME)_linux_$(arch) main.go;)

windows:
	@echo "Building for Windows"
	@mkdir -p $(GODIST)/windows
	$(foreach arch, $(GOARCH), GOOS=windows GOARCH=$(arch) $(GOBUILD) -o $(GODIST)/windows/$(APPNAME)_windows_$(arch).exe main.go;)

darwin:
	@echo "Building for macOS"
	@mkdir -p $(GODIST)/darwin
	$(foreach arch, $(GOARCH), GOOS=darwin GOARCH=$(arch) $(GOBUILD) -o $(GODIST)/darwin/$(APPNAME)_darwin_$(arch) main.go;)

zip:
	@echo "Zipping binaries"
	@cd $(GODIST)/linux && $(ZIP) binaries_linux.zip *
#	@cd $(GODIST)/windows && $(ZIP) binaries_windows.zip *
	@cd $(GODIST)/darwin && $(ZIP) binaries_darwin.zip *
	@$(ZIP) $(GODIST)/binaries.zip $(GODIST)/linux/binaries_linux.zip $(GODIST)/windows/binaries_windows.zip $(GODIST)/darwin/binaries_darwin.zip

clean:
	@echo "Cleaning up"
	@rm -rf $(GODIST)

.PHONY: clean