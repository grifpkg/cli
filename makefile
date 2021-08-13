# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_NAME=grif
BINARY_UNIX=$(BINARY_NAME)_unix

all: deps build
build:
		cd ./src && env GOOS="windows" GOARCH="amd64" $(GOBUILD) -tags netgo -a -o "./../target/$(BINARY_NAME) windows (x64).exe" -v .
		cd ./src && env GOOS="windows" GOARCH="386" $(GOBUILD) -tags netgo -a -o "./../target/$(BINARY_NAME) windows (x32).exe" -v .
		cd ./src && env GOOS="linux" GOARCH="amd64" $(GOBUILD) -tags netgo -a -o "./../target/$(BINARY_NAME) linux (x64)" -v .
		cd ./src && env GOOS="linux" GOARCH="arm64"  $(GOBUILD) -tags netgo -a -o "./../target/$(BINARY_NAME) linux (x64 arm)" -v .
		cd ./src && env GOOS="linux" GOARCH="386" $(GOBUILD) -tags netgo -a -o "./../target/$(BINARY_NAME) linux (x32)" -v .
		cd ./src && env GOOS="openbsd" GOARCH="amd64" $(GOBUILD) -tags netgo -a -o "./../target/$(BINARY_NAME) openbsd (x64)" -v .
		cd ./src && env GOOS="openbsd" GOARCH="arm64" $(GOBUILD) -tags netgo -a -o "./../target/$(BINARY_NAME) openbsd (x64 arm)" -v .
		cd ./src && env GOOS="openbsd" GOARCH="386" $(GOBUILD) -tags netgo -a -o "./../target/$(BINARY_NAME) openbsd (x32)" -v .
		cd ./src && env GOOS="darwin" GOARCH="amd64" $(GOBUILD) -tags netgo -a -o "./../target/$(BINARY_NAME) macos (x64)" -v .
		cd ./src && env GOOS="darwin" GOARCH="arm64" $(GOBUILD) -tags netgo -a -o "./../target/$(BINARY_NAME) macos (x64 arm)" -v .
clean:
		$(GOCLEAN)
		rm -r ./target/
deps:
		cd ./src && go mod tidy