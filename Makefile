.PHONY: vendor

include bin/build/make/go.mak

# Encode a config.
encode-config:
	cat test/$(kind).yml | base64
