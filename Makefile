OUTPUT=./_build
SRC=$(shell find . -iname "*.go")
LDFLAGS='-X main.pkgType=binary -s -w'
RESOURCES=$(wildcard ./console/resources/*.html)

all: binaries msi deb

msi:
	make $(OUTPUT)/lean-cli-setup-x86.msi
	make $(OUTPUT)/lean-cli-setup-x64.msi

$(OUTPUT)/lean-cli-setup-x86.msi: $(OUTPUT)/lean-windows-x86.exe
	wixl -a x86 packaging/msi/lean-cli-x86.wxs -o $@

$(OUTPUT)/lean-cli-setup-x64.msi: $(OUTPUT)/lean-windows-x64.exe
	wixl -a x64 packaging/msi/lean-cli-x64.wxs -o $@

deb:
	make $(OUTPUT)/lean-cli-x86.deb
	make $(OUTPUT)/lean-cli-x64.deb

$(OUTPUT)/lean-cli-x86.deb: $(OUTPUT)/lean-linux-x86
	mkdir -p $(OUTPUT)/x86-deb/DEBIAN/
	mkdir -p $(OUTPUT)/x86-deb/usr/bin/
	cp $(OUTPUT)/lean-linux-x86 $(OUTPUT)/x86-deb/usr/bin/lean
	cp packaging/deb/control-x86 $(OUTPUT)/x86-deb/DEBIAN/control
	dpkg-deb --build $(OUTPUT)/x86-deb $@
	rm -rf $(OUTPUT)/x86-deb

$(OUTPUT)/lean-cli-x64.deb: $(OUTPUT)/lean-linux-x64
	mkdir -p $(OUTPUT)/x64-deb/DEBIAN/
	mkdir -p $(OUTPUT)/x64-deb/usr/bin/
	cp $(OUTPUT)/lean-linux-x64 $(OUTPUT)/x64-deb/usr/bin/lean
	cp packaging/deb/control-x64 $(OUTPUT)/x64-deb/DEBIAN/control
	dpkg-deb --build $(OUTPUT)/x64-deb $@
	rm -rf $(OUTPUT)/x64-deb

binaries: $(SRC)
	make $(OUTPUT)/lean-windows-x86.exe
	make $(OUTPUT)/lean-windows-x64.exe
	make $(OUTPUT)/lean-macos-x64
	make $(OUTPUT)/lean-linux-x86
	make $(OUTPUT)/lean-linux-x64

$(OUTPUT)/lean-windows-x86.exe: $(SRC) resources
	GOOS=windows GOARCH=386 go build -o $@ -ldflags=$(LDFLAGS) github.com/leancloud/lean-cli/lean

$(OUTPUT)/lean-windows-x64.exe: $(SRC) resources
	GOOS=windows GOARCH=amd64 go build -o $@ -ldflags=$(LDFLAGS) github.com/leancloud/lean-cli/lean

$(OUTPUT)/lean-macos-x64: $(SRC) resources
	GOOS=darwin GOARCH=amd64 go build -o $@ -ldflags=$(LDFLAGS) github.com/leancloud/lean-cli/lean

$(OUTPUT)/lean-linux-x86: $(SRC) resources
	GOOS=linux GOARCH=386 go build -o $@ -ldflags=$(LDFLAGS) github.com/leancloud/lean-cli/lean

$(OUTPUT)/lean-linux-x64: $(SRC) resources
	GOOS=linux GOARCH=amd64 go build -o $@ -ldflags=$(LDFLAGS) github.com/leancloud/lean-cli/lean

install: resources
	GOOS=$(GOOS) go install github.com/leancloud/lean-cli/lean

test:
	sh test.sh

resources:
	(cd console; $(MAKE))

clean:
	rm -rf $(OUTPUT)

.PHONY: test msi deb install clean resources
