default: test

build:
    go build ./...

test:
    go test ./...

vet:
    go vet ./...

install:
    @echo "axon-agent is a library, nothing to install"
