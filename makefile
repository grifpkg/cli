# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_NAME=grif
BINARY_UNIX=$(BINARY_NAME)_unix

all: deps build
build:
		cd ./src && env GOOS="windows" GOARCH="amd64" $(GOBUILD) -tags netgo -a -o "./../target/windows.x64.$(BINARY_NAME).exe" -v .
		cd ./src && env GOOS="windows" GOARCH="386" $(GOBUILD) -tags netgo -a -o "./../target/windows.x32.$(BINARY_NAME).exe" -v .
		cd ./src && env GOOS="linux" GOARCH="amd64" $(GOBUILD) -tags netgo -a -o "./../target/linux.x64.$(BINARY_NAME)" -v .
		cd ./src && env GOOS="linux" GOARCH="arm64"  $(GOBUILD) -tags netgo -a -o "./../target/linux.x64.arm.$(BINARY_NAME)" -v .
		cd ./src && env GOOS="linux" GOARCH="386" $(GOBUILD) -tags netgo -a -o "./../target/linux.x32.$(BINARY_NAME)" -v .
		cd ./src && env GOOS="openbsd" GOARCH="amd64" $(GOBUILD) -tags netgo -a -o "./../target/openbsd.x64.$(BINARY_NAME)" -v .
		cd ./src && env GOOS="openbsd" GOARCH="arm64" $(GOBUILD) -tags netgo -a -o "./../target/openbsd.x64.arm.$(BINARY_NAME)" -v .
		cd ./src && env GOOS="openbsd" GOARCH="386" $(GOBUILD) -tags netgo -a -o "./../target/openbsd.x32.$(BINARY_NAME)" -v .
		cd ./src && env GOOS="darwin" GOARCH="amd64" $(GOBUILD) -tags netgo -a -o "./../target/macos.x64.$(BINARY_NAME)" -v .
		cd ./src && env GOOS="darwin" GOARCH="arm64" $(GOBUILD) -tags netgo -a -o "./../target/macos.x64.arm.$(BINARY_NAME)" -v .
clean:
		$(GOCLEAN)
		rm -r ./target/
deps:
		cd ./src && go mod tidy