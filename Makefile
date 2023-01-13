.PHONY: vendor

include bin/build/make/go.mak

# Run all the specs.
specs:
	go test -race -mod vendor -v -covermode=atomic -coverpkg=./... -coverprofile=test/profile.cov ./...

# Send coveralls data.
goveralls: remove-generated-coverage
	goveralls -coverprofile=test/final.cov -service=circle-ci -repotoken=gppTG6I7O5tni1mg336nEm8DFRXwANhJV

# Run security checks.
sec:
	gosec -quiet -exclude-dir=test -exclude=G104 ./...

# Start the environment.
start:
	bin/build/docker/env start

# Stop the environment.
stop:
	bin/build/docker/env stop
