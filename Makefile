NAME?=vault-artifactory-secrets-plugin

get:
	go get ./...

test:
	go test -v ./...

lint:
	go list ./... | xargs go vet -tags testing
	go list ./... | xargs golint

build:
	go build -o plugins/$(NAME)

dev-server:
	vault server -log-level=debug -dev -dev-root-token-id=root -dev-plugin-dir=./plugins

