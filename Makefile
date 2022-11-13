.PHONY: vendor

# Setup everything
setup: dep

download:
	go mod download

tidy:
	go mod tidy

vendor:
	go mod vendor

get:
	go get $(module)

get-all:
	go get -u all

# Setup go deps
dep: download tidy vendor

# Lint all the go code
lint:
	golangci-lint run --timeout 5m

# Fix the lint issues in the go code (if possible)
fix-lint:
	golangci-lint run --timeout 5m --fix

# Run all the specs
specs:
	go test -race -mod vendor -v -covermode=atomic -coverpkg=./... -coverprofile=test/profile.cov ./...

remove-generated-coverage:
	cat test/profile.cov | grep -v "test" > test/final.cov

# Get the HTML coverage for go
html-coverage: remove-generated-coverage
	go tool cover -html test/final.cov

# Get the func coverage for go
func-coverage: remove-generated-coverage
	go tool cover -func test/final.cov

# Send coveralls data
goveralls: remove-generated-coverage
	goveralls -coverprofile=test/final.cov -service=circle-ci -repotoken=gppTG6I7O5tni1mg336nEm8DFRXwANhJV

# Check outdated go deps
outdated:
	go list -u -m -mod=mod -json all | go-mod-outdated -update -direct

# Update go dep
update-dep: get tidy vendor

# Update all go dep
update-all-deps: get-all tidy vendor

# Run security checks.
sec:
	gosec -quiet -exclude-dir=test -exclude=G104 ./...

# Start the environment
start:
	tools/env start

# Stop the environment
stop:
	tools/env stop
