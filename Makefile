BINARY := dmp
PACKAGE := ./cmd/dmp

.PHONY: test build clean

test:
	go test ./...

build:
	go build -o $(BINARY) $(PACKAGE)

clean:
	rm -f $(BINARY)
