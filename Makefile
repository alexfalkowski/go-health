fuzztime ?= 1000x

include bin/build/make/help.mak
include bin/build/make/go.mak
include bin/build/make/git.mak
include bin/build/make/claude.mak
include bin/build/make/codex.mak

# Run all the benchmarks.
benchmarks: server-benchmarks

server-benchmarks:
	@$(MAKE) package=server benchtime=100x benchmark

# Run bounded fuzz smoke tests. Set fuzztime=<count>x to override the default 1000x per target.
fuzzes: checker-fuzz subscriber-fuzz server-fuzz

checker-fuzz:
	@$(MAKE) package=checker name=FuzzHTTPCheckerRequestAndStatus fuzz
	@$(MAKE) package=checker name=FuzzOnlineCheckerURLsAndStatuses fuzz

subscriber-fuzz:
	@$(MAKE) package=subscriber name=FuzzErrorsAggregationAndCopy fuzz

server-fuzz:
	@$(MAKE) package=server name=FuzzServiceObserveValidation fuzz
