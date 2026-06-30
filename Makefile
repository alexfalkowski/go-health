REQUESTED_PACKAGE := $(value package)

include bin/build/make/help.mak
include bin/build/make/go.mak
include bin/build/make/git.mak

ifeq ($(REQUESTED_PACKAGE),)
benchmark benchmark-pprof: package = server
endif
