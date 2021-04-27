.PHONY: vendor

setup-outdated:
	go get -u github.com/psampaz/go-mod-outdated

setup: setup-outdated tidy

lint: dep
	golangci-lint run

fix-lint: dep
	golangci-lint run --fix

specs: dep
	go test -race -mod vendor -v -covermode=atomic -coverpkg=./... -coverprofile=test/profile.cov ./...

coverage:
	go tool cover -html=test/profile.cov

download:
	go mod download

tidy:
	go mod tidy

vendor:
	go mod vendor

dep: download tidy vendor

outdated:
	go list -u -m -mod=mod -json all | go-mod-outdated -update -direct

goveralls:
	goveralls -coverprofile=test/profile.cov -service=circle-ci -repotoken=gppTG6I7O5tni1mg336nEm8DFRXwANhJV
