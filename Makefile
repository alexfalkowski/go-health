.PHONY: vendor

setup-outdated:
	go get -u github.com/psampaz/go-mod-outdated

setup: setup-outdated tidy

lint: dep
	golangci-lint run

fix-lint: dep
	golangci-lint run --fix

specs: dep
	go test -mod vendor -v -covermode=count -coverpkg=./... -coverprofile=rpt/profile.cov ./...

coverage:
	go tool cover -html=rpt/profile.cov

download:
	go mod download

tidy:
	go mod tidy

vendor:
	go mod vendor

dep: download tidy vendor

outdated:
	go list -u -m -mod=mod -json all | go-mod-outdated -update -direct
