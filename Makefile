NAME?=vault-artifactory-secrets-plugin

get:
	go get ./...

test:
	go test -short -parallel=10 -v -covermode=count -coverprofile=coverage.out ./...

report: .tools/gocover-cobertura
	go tool cover -html=coverage.out -o coverage.html
	.tools/gocover-cobertura < coverage.out > coverage.xml

lint: .tools/golangci-lint
	.tools/golangci-lint run

build:
	go build -o plugins/$(NAME)

dev-server:
	vault server -log-level=debug -dev -dev-root-token-id=root -dev-plugin-dir=./plugins

.tools/golangci-lint:
	export GOBIN=$(shell pwd)/.tools; go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.39.0

.tools/gocover-cobertura:
	export GOBIN=$(shell pwd)/.tools; go install github.com/boumenot/gocover-cobertura
