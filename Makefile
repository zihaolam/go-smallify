GO := go

BINARY_NAME := bin/smallify

.PHONY: all build run clean

build:
	$(GO) build -o $(BINARY_NAME) *.go

run: build
	./$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)