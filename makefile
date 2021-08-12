    # Go parameters
    GOCMD=go
    GOBUILD=$(GOCMD) build
    GOCLEAN=$(GOCMD) clean
    GOGET=$(GOCMD) get
    BINARY_NAME=grif
    BINARY_UNIX=$(BINARY_NAME)_unix
    
    all: deps build
    build: 
			env GOOS="windows" GOARCH="amd64" $(GOBUILD) -o "./target/windows.x64.$(BINARY_NAME).exe" -v ./src/
			env GOOS="windows" GOARCH="386" $(GOBUILD) -o "./target/windows.x32.$(BINARY_NAME).exe" -v ./src/
			env GOOS="linux" GOARCH="amd64" $(GOBUILD) -o "./target/linux.x64.$(BINARY_NAME)" -v ./src/
			env GOOS="linux" GOARCH="arm64" $(GOBUILD) -o "./target/linux.x64.arm.$(BINARY_NAME)" -v ./src/
			env GOOS="linux" GOARCH="386" $(GOBUILD) -o "./target/linux.x32.$(BINARY_NAME)" -v ./src/
			env GOOS="openbsd" GOARCH="amd64" $(GOBUILD) -o "./target/openbsd.x64.$(BINARY_NAME)" -v ./src/
			env GOOS="openbsd" GOARCH="arm64" $(GOBUILD) -o "./target/openbsd.x64.arm.$(BINARY_NAME)" -v ./src/
			env GOOS="openbsd" GOARCH="386" $(GOBUILD) -o "./target/openbsd.x32.$(BINARY_NAME)" -v ./src/
			env GOOS="darwin" GOARCH="amd64" $(GOBUILD) -o "./target/macos.x64.$(BINARY_NAME)" -v ./src/
			env GOOS="darwin" GOARCH="arm64" $(GOBUILD) -o "./target/macos.x64.arm.$(BINARY_NAME)" -v ./src/
    clean: 
			$(GOCLEAN)
			rm -r ./target/
    deps:
			$(GOGET) "github.com/fatih/color"
			$(GOGET) "github.com/spf13/cobra"
			$(GOGET) "github.com/AlecAivazis/survey/v2"
			${GOGET} "github.com/segmentio/ksuid"
			${GOGET} "github.com/inconshreveable/mousetrap"