OUTPUT=./build
SRC=$(shell find lean/ -iname "*.go")

all:
	GOOS=darwin GOARCH=amd64 make $(OUTPUT)/lean-darwin-amd64
	GOOS=windows GOARCH=386 make $(OUTPUT)/lean-windows-386
	GOOS=windows GOARCH=amd64 make $(OUTPUT)/lean-windows-amd64

$(OUTPUT)/lean-$(GOOS)-$(GOARCH): $(SRC)
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -x -v -o $@ github.com/leancloud/lean-cli/lean

install:
	GOOS=$(GOOS) go install -x -v github.com/leancloud/lean-cli/lean

test:
	go test -v github.com/leancloud/lean-cli/lean/boilerplate
	go test -v github.com/leancloud/lean-cli/lean/console
	go test -v github.com/leancloud/lean-cli/lean/apps

clean:
	rm -rf $(OUTPUT)

.PHONY: test
