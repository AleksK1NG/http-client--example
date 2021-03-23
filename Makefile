.PHONY:


# ==============================================================================
# Run
URL = https://someaddress.com
run:
	@echo "Starting http client 🚀"
	go run cmd/main.go -url=$(URL)

# ==============================================================================
# Linters

run-linter:
	echo "Starting linters"
	golangci-lint run ./...


# ==============================================================================
# Modules support

deps-reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

tidy:
	go mod tidy
	go mod vendor

deps-upgrade:
	# go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)
	go get -u -t -d -v ./...
	go mod tidy
	go mod vendor

deps-cleancache:
	go clean -modcache

list:
	go list -mod=mod all