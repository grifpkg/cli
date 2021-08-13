# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_NAME=grif
BINARY_UNIX=$(BINARY_NAME)_unix

all: deps build
build:
		cd ./src && env GOOS="windows" GOARCH="amd64" $(GOBUILD) -tags netgo -a -o "./../target/$(BINARY_NAME)_windows_x64.exe" -v .
		cd ./src && env GOOS="windows" GOARCH="386" $(GOBUILD) -tags netgo -a -o "./../target/$(BINARY_NAME)_windows_x32.exe" -v .
		cd ./src && env GOOS="linux" GOARCH="amd64" $(GOBUILD) -tags netgo -a -o "./../target/$(BINARY_NAME)_linux_x64" -v .
		cd ./src && env GOOS="linux" GOARCH="arm64"  $(GOBUILD) -tags netgo -a -o "./../target/$(BINARY_NAME)_linux_x64_arm" -v .
		cd ./src && env GOOS="linux" GOARCH="386" $(GOBUILD) -tags netgo -a -o "./../target/$(BINARY_NAME)_linux_x32" -v .
		cd ./src && env GOOS="openbsd" GOARCH="amd64" $(GOBUILD) -tags netgo -a -o "./../target/$(BINARY_NAME)_openbsd_x64" -v .
		cd ./src && env GOOS="openbsd" GOARCH="arm64" $(GOBUILD) -tags netgo -a -o "./../target/$(BINARY_NAME)_openbsd_x64_arm" -v .
		cd ./src && env GOOS="openbsd" GOARCH="386" $(GOBUILD) -tags netgo -a -o "./../target/$(BINARY_NAME)_openbsd_x32" -v .
		cd ./src && env GOOS="darwin" GOARCH="amd64" $(GOBUILD) -tags netgo -a -o "./../target/$(BINARY_NAME)_macos_x64" -v .
		cd ./src && env GOOS="darwin" GOARCH="arm64" $(GOBUILD) -tags netgo -a -o "./../target/$(BINARY_NAME)_macos_x64_arm" -v .
clean:
		$(GOCLEAN)
		rm -r ./target/
deps:
		cd ./src && go mod tidy