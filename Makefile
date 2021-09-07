OUTPUT=./_build
SRC=$(shell find . -iname "*.go")
LDFLAGS='-X main.pkgType=binary -s -w'
LDFLAGS_TDS="-X main.pkgType=binary -X github.com/leancloud/lean-cli/version.Distribution=tds -s -w"

all: binaries msi deb

msi:
	make $(OUTPUT)/lean-cli-setup-x86.msi
	make $(OUTPUT)/lean-cli-setup-x64.msi
	make $(OUTPUT)/tds-cli-setup-x86.msi
	make $(OUTPUT)/tds-cli-setup-x64.msi

$(OUTPUT)/lean-cli-setup-x86.msi: $(OUTPUT)/lean-windows-x86.exe
	wixl -a x86 packaging/msi/lean-cli-x86.wxs -o $@

$(OUTPUT)/lean-cli-setup-x64.msi: $(OUTPUT)/lean-windows-x64.exe
	wixl -a x64 packaging/msi/lean-cli-x64.wxs -o $@

$(OUTPUT)/tds-cli-setup-x86.msi: $(OUTPUT)/tds-windows-x86.exe
	wixl -a x86 packaging/msi/tds-cli-x86.wxs -o $@

$(OUTPUT)/tds-cli-setup-x64.msi: $(OUTPUT)/tds-windows-x64.exe
	wixl -a x64 packaging/msi/tds-cli-x64.wxs -o $@

deb:
	make $(OUTPUT)/lean-cli-x86.deb
	make $(OUTPUT)/lean-cli-x64.deb
	make $(OUTPUT)/tds-cli-x86.deb
	make $(OUTPUT)/tds-cli-x64.deb

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

$(OUTPUT)/tds-cli-x86.deb: $(OUTPUT)/tds-linux-x86
	mkdir -p $(OUTPUT)/x86-deb/DEBIAN/
	mkdir -p $(OUTPUT)/x86-deb/usr/bin/
	cp $(OUTPUT)/tds-linux-x86 $(OUTPUT)/x86-deb/usr/bin/tds
	cp packaging/deb/control-x86 $(OUTPUT)/x86-deb/DEBIAN/control
	dpkg-deb --build $(OUTPUT)/x86-deb $@
	rm -rf $(OUTPUT)/x86-deb

$(OUTPUT)/tds-cli-x64.deb: $(OUTPUT)/tds-linux-x64
	mkdir -p $(OUTPUT)/x64-deb/DEBIAN/
	mkdir -p $(OUTPUT)/x64-deb/usr/bin/
	cp $(OUTPUT)/tds-linux-x64 $(OUTPUT)/x64-deb/usr/bin/tds
	cp packaging/deb/control-x64 $(OUTPUT)/x64-deb/DEBIAN/control
	dpkg-deb --build $(OUTPUT)/x64-deb $@
	rm -rf $(OUTPUT)/x64-deb

binaries: $(SRC)
	make $(OUTPUT)/lean-windows-x86.exe
	make $(OUTPUT)/lean-windows-x64.exe
	make $(OUTPUT)/lean-macos-x64
	make $(OUTPUT)/lean-macos-arm64
	make $(OUTPUT)/lean-linux-x86
	make $(OUTPUT)/lean-linux-x64
	make $(OUTPUT)/tds-windows-x86.exe
	make $(OUTPUT)/tds-windows-x64.exe
	make $(OUTPUT)/tds-macos-x64
	make $(OUTPUT)/tds-macos-arm64
	make $(OUTPUT)/tds-linux-x86
	make $(OUTPUT)/tds-linux-x64

$(OUTPUT)/lean-windows-x86.exe: $(SRC)
	GOOS=windows GOARCH=386 go build -o $@ -ldflags=$(LDFLAGS) github.com/leancloud/lean-cli/lean

$(OUTPUT)/lean-windows-x64.exe: $(SRC)
	GOOS=windows GOARCH=amd64 go build -o $@ -ldflags=$(LDFLAGS) github.com/leancloud/lean-cli/lean

$(OUTPUT)/lean-macos-x64: $(SRC)
	GOOS=darwin GOARCH=amd64 go build -o $@ -ldflags=$(LDFLAGS) github.com/leancloud/lean-cli/lean

$(OUTPUT)/lean-macos-arm64: $(SRC)
	GOOS=darwin GOARCH=arm64 go build -o $@ -ldflags=$(LDFLAGS) github.com/leancloud/lean-cli/lean

$(OUTPUT)/lean-linux-x86: $(SRC)
	GOOS=linux GOARCH=386 go build -o $@ -ldflags=$(LDFLAGS) github.com/leancloud/lean-cli/lean

$(OUTPUT)/lean-linux-x64: $(SRC)
	GOOS=linux GOARCH=amd64 go build -o $@ -ldflags=$(LDFLAGS) github.com/leancloud/lean-cli/lean

$(OUTPUT)/tds-windows-x86.exe: $(SRC)
	GOOS=windows GOARCH=386 go build -o $@ -ldflags=$(LDFLAGS_TDS) github.com/leancloud/lean-cli/lean

$(OUTPUT)/tds-windows-x64.exe: $(SRC)
	GOOS=windows GOARCH=amd64 go build -o $@ -ldflags=$(LDFLAGS_TDS) github.com/leancloud/lean-cli/lean

$(OUTPUT)/tds-macos-x64: $(SRC)
	GOOS=darwin GOARCH=amd64 go build -o $@ -ldflags=$(LDFLAGS_TDS) github.com/leancloud/lean-cli/lean

$(OUTPUT)/tds-macos-arm64: $(SRC)
	GOOS=darwin GOARCH=arm64 go build -o $@ -ldflags=$(LDFLAGS_TDS) github.com/leancloud/lean-cli/lean

$(OUTPUT)/tds-linux-x86: $(SRC)
	GOOS=linux GOARCH=386 go build -o $@ -ldflags=$(LDFLAGS_TDS) github.com/leancloud/lean-cli/lean

$(OUTPUT)/tds-linux-x64: $(SRC)
	GOOS=linux GOARCH=amd64 go build -o $@ -ldflags=$(LDFLAGS_TDS) github.com/leancloud/lean-cli/lean

install:
	GOOS=$(GOOS) go install github.com/leancloud/lean-cli/lean

test:
	go test github.com/leancloud/lean-cli/lean -v

clean:
	rm -rf $(OUTPUT)

.PHONY: test msi deb install clean
