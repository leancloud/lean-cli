OUTPUT=./_build
SRC=$(shell find . -iname "*.go")
LDFLAGS='-X main.pkgType="binary" -s -w'
RESOURCES=$(wildcard ./console/resources/*.html)

all: binaries msi deb

binaries: $(SRC)
	GOOS=darwin GOARCH=amd64 POSTFIX= make $(OUTPUT)/lean-darwin-amd64
	GOOS=windows GOARCH=386 POSTFIX=.exe make $(OUTPUT)/lean-windows-386.exe
	GOOS=windows GOARCH=amd64 POSTFIX=.exe make $(OUTPUT)/lean-windows-amd64.exe
	GOOS=linux GOARCH=amd64 POSTFIX= make $(OUTPUT)/lean-linux-amd64
	GOOS=linux GOARCH=386 POSTFIX= make $(OUTPUT)/lean-linux-386

msi:
	wixl -a x86 packaging/msi/lean-cli-x86.wxs -o $(OUTPUT)/lean-cli-setup-x86.msi
	wixl -a x64 packaging/msi/lean-cli-x64.wxs -o $(OUTPUT)/lean-cli-setup-x64.msi

deb:
	mkdir -p $(OUTPUT)/x86-deb/DEBIAN/
	mkdir -p $(OUTPUT)/x86-deb/usr/bin/
	cp $(OUTPUT)/lean-linux-386 $(OUTPUT)/x86-deb/usr/bin/lean
	cp packaging/deb/control-x86 $(OUTPUT)/x86-deb/DEBIAN/control
	dpkg-deb --build $(OUTPUT)/x86-deb $(OUTPUT)/lean-cli-x86.deb
	mkdir -p $(OUTPUT)/x64-deb/DEBIAN/
	mkdir -p $(OUTPUT)/x64-deb/usr/bin/
	cp $(OUTPUT)/lean-linux-amd64 $(OUTPUT)/x64-deb/usr/bin/lean
	cp packaging/deb/control-x64 $(OUTPUT)/x64-deb/DEBIAN/control
	dpkg-deb --build $(OUTPUT)/x64-deb $(OUTPUT)/lean-cli-x64.deb

$(OUTPUT)/lean-$(GOOS)-$(GOARCH)$(POSTFIX): $(SRC)
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $@ -ldflags=$(LDFLAGS) github.com/leancloud/lean-cli/lean

install: resources
	GOOS=$(GOOS) go install github.com/leancloud/lean-cli/lean

test:
	go test -v github.com/leancloud/lean-cli/boilerplate
	go test -v github.com/leancloud/lean-cli/console
	go test -v github.com/leancloud/lean-cli/apps
	go test -v github.com/leancloud/lean-cli/stats

resources:
	(cd console; $(MAKE))

clean:
	rm -rf $(OUTPUT)

.PHONY: test msi deb install clean resources
