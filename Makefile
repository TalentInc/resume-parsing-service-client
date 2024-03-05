.PHONY: help
## help: shows this help message
help:
	@ echo "Usage: make [target]\n"
	@ sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: test
## test: run unit tests
test:
	@ go test -v ./... -count=1

.PHONY: coverage
## coverage: run unit tests and generate coverage report in html format
coverage:
	@ go test -coverprofile=coverage.out ./...  && go tool cover -html=coverage.out

.PHONY: lint
## lint: runs linter for all packages
lint: 
	@ echo "Running linter..."
	@ docker run  --rm -v "`pwd`:/workspace:cached" -w "/workspace/." golangci/golangci-lint:latest golangci-lint run

.PHONY: vul-setup
## vul-setup: installs Golang's vulnerability check tool
vul-setup:
	@ if [ -z "$$(which govulncheck)" ]; then echo "Installing Golang's vulnerability detection tool..."; go install golang.org/x/vuln/cmd/govulncheck@latest; fi

.PHONY: vul-check
## vul-check: checks for any known vulnerabilities
vul-check: vul-setup
	@ @ echo "Checking for any known vulnerabilities..."
	@ govulncheck ./...