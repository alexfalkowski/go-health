include bin/build/make/help.mak
include bin/build/make/go.mak
include bin/build/make/git.mak

# Run all the benchmarks.
benchmarks: server-benchmarks

server-benchmarks:
	@$(MAKE) package=server benchtime=100x benchmark

# Run bounded fuzz smoke tests. Set fuzztime=<duration> to override the default 1s per target.
fuzzes: checker-fuzz subscriber-fuzz server-fuzz

checker-fuzz:
	@$(MAKE) package=checker name=FuzzHTTPCheckerRequestAndStatus fuzztime=$(or $(fuzztime),1s) fuzz
	@$(MAKE) package=checker name=FuzzOnlineCheckerURLsAndStatuses fuzztime=$(or $(fuzztime),1s) fuzz

subscriber-fuzz:
	@$(MAKE) package=subscriber name=FuzzErrorsAggregationAndCopy fuzztime=$(or $(fuzztime),1s) fuzz

server-fuzz:
	@$(MAKE) package=server name=FuzzServiceObserveValidation fuzztime=$(or $(fuzztime),1s) fuzz
