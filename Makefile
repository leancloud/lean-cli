OUTPUT=./build
SRC=$(shell find lean/ -iname "*.go")
LDFLAGS='-X main.pkgType="binary" -s -w'

binaries:
	GOOS=darwin GOARCH=amd64 POSTFIX= make $(OUTPUT)/lean-darwin-amd64
	GOOS=windows GOARCH=386 POSTFIX=.exe make $(OUTPUT)/lean-windows-386.exe
	GOOS=windows GOARCH=amd64 POSTFIX=.exe make $(OUTPUT)/lean-windows-amd64.exe
	GOOS=linux GOARCH=amd64 POSTFIX= make $(OUTPUT)/lean-linux-amd64
	GOOS=linux GOARCH=386 POSTFIX= make $(OUTPUT)/lean-linux-386

$(OUTPUT)/lean-$(GOOS)-$(GOARCH)$(POSTFIX): $(SRC)
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $@ -ldflags=$(LDFLAGS) github.com/leancloud/lean-cli/lean

install:
	GOOS=$(GOOS) go install github.com/leancloud/lean-cli/lean

test:
	go test -v github.com/leancloud/lean-cli/lean/boilerplate
	go test -v github.com/leancloud/lean-cli/lean/console
	go test -v github.com/leancloud/lean-cli/lean/apps
	go test -v github.com/leancloud/lean-cli/lean/stats

clean:
	rm -rf $(OUTPUT)

.PHONY: test
