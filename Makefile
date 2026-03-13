BIN      := vimtutor
LDFLAGS  := -ldflags "-s -w"

.PHONY: build run test clean install

build:
	go build $(LDFLAGS) -o bin/$(BIN) .

run:
	go run . learn

test:
	go test ./...

clean:
	rm -rf bin/

install:
	go install $(LDFLAGS) .
