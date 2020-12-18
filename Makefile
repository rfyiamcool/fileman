VERSION :=v0.0.1

.PHONY: build install release release-linux release-mac

run:
	@go build .
	@./fileman

build:
	@go build -ldflags '-w -s' -o fileman

install: build
	@cp -f fileman /usr/local/bin/

release: release-linux release-mac

release-linux:
	@env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w -s' -o fileman
	@tar zcvf fileman-$(VERSION)-linux-amd64.tar.gz ./fileman
	@rm -f fileman

release-mac:
	@env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags '-w -s' -o fileman
	@tar zcvf fileman-$(VERSION)-darwin-amd64.tar.gz ./fileman
	@rm -f fileman