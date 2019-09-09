# This makefile is for GNU Make, and nowadays provided
# just for compatibility and preservation of traditions.
# Please use CMake in case of any difficulties or problems.
#
# Preprocessor macros (for MDBX_OPTIONS) of interest...
# Note that the defaults should already be correct for most platforms;
# you should not need to change any of these. Read their descriptions
# in README and source code if you do. There may be other macros of interest.

# install sandbox
SANDBOX ?=

# install prefixes (inside sandbox)
prefix  ?= /usr/local
mandir  ?= $(prefix)/man

# lib/bin suffix for multiarch/biarch, e.g. '.x86_64'
suffix  ?=

CC      ?= gcc
CXX     ?= g++
LD      ?= ld
MDBX_OPTIONS ?= -D_GNU_SOURCE=1 -DNDEBUG=1
CFLAGS  ?= -O2 -g3 -Wall -Werror -Wextra -ffunction-sections -fPIC -fvisibility=hidden -std=gnu11 -pthread
CXXFLAGS = -std=c++11 $(filter-out -std=gnu11,$(CFLAGS))

# LY: '--no-as-needed,-lrt' for ability to built with modern glibc, but then run with the old
LDFLAGS ?= $(shell $(LD) --help 2>/dev/null | grep -q -- --gc-sections && echo '-Wl,--gc-sections,-z,relro,-O1')$(shell $(LD) --help 2>/dev/null | grep -q -- -dead_strip && echo '-Wl,-dead_strip')
EXE_LDFLAGS ?= -pthread

################################################################################

UNAME      := $(shell uname -s 2>/dev/null || echo Unknown)
define uname2sosuffix
  case "$(UNAME)" in
    Darwin*|Mach*) echo dylib;;
    CYGWIN*|MINGW*|MSYS*|Windows*) echo dll;;
    *) echo so;;
  esac
endef
SO_SUFFIX  := $(shell $(uname2sosuffix))

HEADERS    := mdbx.h
LIBRARIES  := libmdbx.a libmdbx.$(SO_SUFFIX)
TOOLS      := mdbx_stat mdbx_copy mdbx_dump mdbx_load mdbx_chk
MANPAGES   := mdbx_stat.1 mdbx_copy.1 mdbx_dump.1 mdbx_load.1
SHELL      := /bin/bash

.PHONY: mdbx all install clean check coverage

all: $(LIBRARIES) $(TOOLS)

mdbx: libmdbx.a libmdbx.$(SO_SUFFIX)

tools: $(TOOLS)

strip: all
	strip libmdbx.$(SO_SUFFIX) $(TOOLS)

install: $(LIBRARIES) $(TOOLS) $(HEADERS)
	mkdir -p $(SANDBOX)$(prefix)/bin$(suffix) \
		&& cp -t $(SANDBOX)$(prefix)/bin$(suffix) $(TOOLS) && \
	mkdir -p $(SANDBOX)$(prefix)/lib$(suffix) \
		&& cp -t $(SANDBOX)$(prefix)/lib$(suffix) $(LIBRARIES) && \
	mkdir -p $(SANDBOX)$(prefix)/include \
		&& cp -t $(SANDBOX)$(prefix)/include $(HEADERS) && \
	mkdir -p $(SANDBOX)$(mandir)/man1 \
		&& cp -t $(SANDBOX)$(mandir)/man1 $(MANPAGES)

clean:
	rm -rf $(TOOLS) mdbx_test @* *.[ao] *.[ls]o *~ tmp.db/* *.gcov *.log *.err src/*.o test/*.o

libmdbx.a: mdbx-static.o
	$(AR) rs $@ $?

libmdbx.$(SO_SUFFIX): mdbx-dylib.o
	$(CC) $(CFLAGS) $^ -pthread -shared $(LDFLAGS) -o $@

#> dist-cutoff-begin
ifeq ($(wildcard mdbx.c),mdbx.c)
#< dist-cutoff-end

################################################################################
# Amalgamated source code, i.e. distributed after `make dists`

mdbx-config.h: mdbx-config.h.in mdbx.c $(lastword $(MAKEFILE_LIST))
	(echo '#define MDBX_BUILD_TIMESTAMP "$(shell date +%Y-%m-%dT%H:%M:%S%z)"' \
	&& echo '#define MDBX_BUILD_OPTIONS_STRING "$(MDBX_OPTIONS)"' \
	&& echo '#define MDBX_BUILD_FLAGS "$(CFLAGS) $(LDFLAGS)"' \
	&& echo '#define MDBX_BUILD_COMPILER "$(shell set -o pipefail; $(CC) --version | head -1 || echo 'Please use GCC or CLANG compatible compiler')"' \
	&& echo '#define MDBX_BUILD_TARGET "$(shell set -o pipefail; LC_ALL=C $(CC) -v 2>&1 | grep -i '^Target:' | cut -d ' ' -f 2- || echo 'Please use GCC or CLANG compatible compiler')"' \
	) > $@ || rm -f $@

mdbx-dylib.o: mdbx-config.h mdbx.c $(lastword $(MAKEFILE_LIST))
	$(CC) $(CFLAGS) $(MDBX_OPTIONS) -DLIBMDBX_EXPORTS=1 -c mdbx.c -o $@

mdbx-static.o: mdbx-config.h mdbx.c $(lastword $(MAKEFILE_LIST))
	$(CC) $(CFLAGS) $(MDBX_OPTIONS) -ULIBMDBX_EXPORTS -c mdbx.c -o $@

mdbx_%:	mdbx_%.c libmdbx.a
	$(CC) $(CFLAGS) $(MDBX_OPTIONS) $^ $(EXE_LDFLAGS) -o $@

#> dist-cutoff-begin
else
################################################################################
# Plain (non-amalgamated) sources

define uname2osal
  case "$(UNAME)" in
    CYGWIN*|MINGW*|MSYS*|Windows*) echo windows;;
    *) echo unix;;
  esac
endef
define uname2titer
  case "$(UNAME)" in
    Darwin*|Mach*) echo 3;;
    *) echo 12;;
  esac
endef
TEST_OSAL  := $(shell $(uname2osal))
TEST_ITER  := $(shell $(uname2titer))
TESTDB     ?= $(shell [ -d /dev/shm ] && echo /dev/shm || echo /tmp)/mdbx-test.db
TESTLOG    ?= $(shell [ -d /dev/shm ] && echo /dev/shm || echo /tmp)/mdbx-test.log
TEST_SRC   := test/osal-$(TEST_OSAL).cc $(filter-out $(wildcard test/osal-*.cc), $(wildcard test/*.cc))
TEST_INC   := $(wildcard test/*.h)
TEST_OBJ   := $(patsubst %.cc,%.o,$(TEST_SRC))

check: all example mdbx_test
	rm -f $(TESTDB) $(TESTLOG) && (set -o pipefail; ./mdbx_test --repeat=$(TEST_ITER) --pathname=$(TESTDB) --dont-cleanup-after basic | tee -a $(TESTLOG) | tail -n 42) \
	&& ./mdbx_chk -vvn $(TESTDB) && ./mdbx_chk -vvn $(TESTDB)-copy

example: mdbx.h tutorial/sample-mdbx.c libmdbx.$(SO_SUFFIX)
	$(CC) $(CFLAGS) -I. tutorial/sample-mdbx.c ./libmdbx.$(SO_SUFFIX) -o example

check-singleprocess: all
	rm -f $(TESTDB) $(TESTLOG) && (set -o pipefail; \
		./mdbx_test --repeat=4 --pathname=$(TESTDB) --dont-cleanup-after --hill && \
		./mdbx_test --repeat=2 --pathname=$(TESTDB) --dont-cleanup-before --dont-cleanup-after --copy \
		| tee -a $(TESTLOG) | tail -n 42) \
	&& ./mdbx_chk -vvn $(TESTDB) && ./mdbx_chk -vvn $(TESTDB)-copy

check-fault: all
	rm -f $(TESTDB) $(TESTLOG) && (set -o pipefail; ./mdbx_test --pathname=$(TESTDB) --inject-writefault=42 --dump-config --dont-cleanup-after basic | tee -a $(TESTLOG) | tail -n 42) \
	; ./mdbx_chk -vvnw $(TESTDB) && ([ ! -e $(TESTDB)-copy ] || ./mdbx_chk -vvn $(TESTDB)-copy)

define test-rule
$(patsubst %.cc,%.o,$(1)): $(1) $(TEST_INC) mdbx.h $(lastword $(MAKEFILE_LIST))
	$(CXX) $(CXXFLAGS) $(MDBX_OPTIONS) -c $(1) -o $$@

endef
$(foreach file,$(TEST_SRC),$(eval $(call test-rule,$(file))))

mdbx_%:	src/tools/mdbx_%.c libmdbx.a
	$(CC) $(CFLAGS) $(MDBX_OPTIONS) $^ $(EXE_LDFLAGS) -o $@

mdbx_test: $(TEST_OBJ) libmdbx.$(SO_SUFFIX)
	$(CXX) $(CXXFLAGS) $(TEST_OBJ) -Wl,-rpath . -L . -l mdbx $(EXE_LDFLAGS) -o $@

ALLOY_DEPS = $(wildcard src/elements/*)
MDBX_VERSION_GIT = ${shell set -o pipefail; git describe --tags | sed -n 's|^v*\([0-9]\{1,\}\.[0-9]\{1,\}\.[0-9]\{1,\}\)\(.*\)|\1|p' || echo 'Please fetch tags and/or install latest git version'}
MDBX_VERSION_SUFFIX = $(shell set -o pipefail; git describe --tags --long --dirty=-dirty | tr -c -s '[a-zA-Z0-9]\n' _ || echo 'Please fetch tags and/or install latest git version')
MDBX_BUILD_SOURCERY = $(shell set -o pipefail; $(MAKE) -s src/elements/version.c && (openssl dgst -r -sha256 src/elements/version.c || sha256sum src/elements/version.c || shasum -a 256 src/elements/version.c) 2>/dev/null | cut -d ' ' -f 1 || echo 'Please install openssl or sha256sum or shasum')_$(MDBX_VERSION_SUFFIX)

src/elements/version.c: src/elements/version.c.in $(lastword $(MAKEFILE_LIST)) .git/HEAD .git/index .git/refs/tags
	sed \
		-e "s|@MDBX_GIT_TIMESTAMP@|$(shell git show --no-patch --format=%cI HEAD || echo 'Please install latest get version')|" \
		-e "s|@MDBX_GIT_TREE@|$(shell git show --no-patch --format=%T HEAD || echo 'Please install latest get version')|" \
		-e "s|@MDBX_GIT_COMMIT@|$(shell git show --no-patch --format=%H HEAD || echo 'Please install latest get version')|" \
		-e "s|@MDBX_GIT_DESCRIBE@|$(shell git describe --tags --long --dirty=-dirty || echo 'Please fetch tags and/or install latest git version')|" \
		-e "s|\$${MDBX_VERSION_MAJOR}|$(shell echo '$(MDBX_VERSION_GIT)' | cut -d . -f 1)|" \
		-e "s|\$${MDBX_VERSION_MINOR}|$(shell echo '$(MDBX_VERSION_GIT)' | cut -d . -f 2)|" \
		-e "s|\$${MDBX_VERSION_RELEASE}|$(shell echo '$(MDBX_VERSION_GIT)' | cut -d . -f 3)|" \
		-e "s|\$${MDBX_VERSION_REVISION}|$(shell git rev-list --count --no-merges HEAD || echo 'Please fetch tags and/or install latest git version')|" \
	src/elements/version.c.in > $@ || rm -f $@

src/elements/config.h: src/elements/version.c $(lastword $(MAKEFILE_LIST))
	(echo '#define MDBX_BUILD_TIMESTAMP "$(shell date +%Y-%m-%dT%H:%M:%S%z)"' \
	&& echo '#define MDBX_BUILD_OPTIONS_STRING "$(MDBX_OPTIONS)"' \
	&& echo '#define MDBX_BUILD_FLAGS "$(CFLAGS) $(LDFLAGS)"' \
	&& echo '#define MDBX_BUILD_COMPILER "$(shell set -o pipefail; $(CC) --version | head -1 || echo 'Please use GCC or CLANG compatible compiler')"' \
	&& echo '#define MDBX_BUILD_TARGET "$(shell set -o pipefail; LC_ALL=C $(CC) -v 2>&1 | grep -i '^Target:' | cut -d ' ' -f 2- || echo 'Please use GCC or CLANG compatible compiler')"' \
	&& echo '#define MDBX_BUILD_SOURCERY $(MDBX_BUILD_SOURCERY)' \
	) > $@ || rm -f $@

mdbx-dylib.o: src/elements/config.h src/elements/version.c src/alloy.c $(ALLOY_DEPS) $(lastword $(MAKEFILE_LIST))
	$(CC) $(CFLAGS) $(MDBX_OPTIONS) -DLIBMDBX_EXPORTS=1 -c src/alloy.c -o $@

mdbx-static.o: src/elements/config.h src/elements/version.c src/alloy.c $(ALLOY_DEPS) $(lastword $(MAKEFILE_LIST))
	$(CC) $(CFLAGS) $(MDBX_OPTIONS) -ULIBMDBX_EXPORTS -c src/alloy.c -o $@

.PHONY: dist
dist: lidmbx-sources-$(MDBX_VERSION_SUFFIX).tar.gz $(lastword $(MAKEFILE_LIST))

lidmbx-sources-$(MDBX_VERSION_SUFFIX).tar.gz: dist/mdbx.c dist/mdbx-config.h.in dist/mdbx.h \
		dist/mdbx_chk.c dist/mdbx_copy.c dist/mdbx_dump.c dist/mdbx_load.c dist/mdbx_stat.c \
		dist/GNUmakefile $(lastword $(MAKEFILE_LIST))
	tar -c -a -f $@ --owner=0 --group=0 -C dist mdbx.c mdbx-config.h.in mdbx.h \
    mdbx_chk.c mdbx_copy.c mdbx_dump.c mdbx_load.c mdbx_stat.c GNUmakefile \
    && rm dist/@tmp-shared_internals.inc

dist/mdbx.h: mdbx.h src/elements/version.c $(lastword $(MAKEFILE_LIST))
	mkdir -p dist && cp $< $@

dist/GNUmakefile: GNUmakefile
	mkdir -p dist && sed -e '/^#> dist-cutoff-begin/,/^#< dist-cutoff-end/d' $< > $@

dist/@tmp-shared_internals.inc: src/elements/version.c $(ALLOY_DEPS) $(lastword $(MAKEFILE_LIST))
	mkdir -p dist && sed \
		-e 's|#include "config.h"|@INCLUDE "mdbx-config.h"\n#define MDBX_ALLOY 1\n#define MDBX_BUILD_SOURCERY $(MDBX_BUILD_SOURCERY)|' \
		-e 's|#include "../../mdbx.h"|@INCLUDE "mdbx.h"|' \
		-e '/#include "defs.h"/r src/elements/defs.h' \
		-e '/#include "osal.h"/r src/elements/osal.h' \
		src/elements/internals.h > $@

dist/mdbx.c: dist/@tmp-shared_internals.inc $(lastword $(MAKEFILE_LIST))
	mkdir -p dist && (cat dist/@tmp-shared_internals.inc \
	&& cat src/elements/core.c src/elements/osal.c src/elements/version.c \
	&& echo '#if defined(_WIN32) || defined(_WIN64)' \
	&& cat src/elements/lck-windows.c && echo '#else /* LCK-implementation */' \
	&& cat src/elements/lck-posix.c && echo '#endif /* LCK-implementation */' \
	) | grep -v -e '#include "' -e '#pragma once' | sed 's|@INCLUDE|#include|' > $@

define dist-tool-rule
dist/mdbx_$(1).c: src/tools/mdbx_$(1).c src/tools/wingetopt.h src/tools/wingetopt.c dist/mdbx.c $(lastword $(MAKEFILE_LIST))
	mkdir -p dist && sed \
		-e '/#include "..\/elements\/internals.h"/r dist/@tmp-shared_internals.inc' \
		-e '/#include "wingetopt.h"/r r/src/tools/wingetopt.c' \
		-e '/#include "wingetopt.h"/r r/src/tools/wingetopt.h' \
		src/tools/mdbx_$(1).c \
	| grep -v -e '#include "' -e '#pragma once' -e '#define MDBX_ALLOY' \
	| sed 's|@INCLUDE|#include|' > $$@

endef
$(foreach file,chk copy dump load stat,$(eval $(call dist-tool-rule,$(file))))

dist/mdbx-config.h.in: src/elements/config.h.in $(lastword $(MAKEFILE_LIST))
	mkdir -p dist && grep -v MDBX_BUILD_SOURCERY $< > $@

endif

################################################################################
# Cross-compilation simple test

CROSS_LIST = mips-linux-gnu-gcc \
	powerpc64-linux-gnu-gcc powerpc-linux-gnu-gcc \
	arm-linux-gnueabihf-gcc aarch64-linux-gnu-gcc

# hppa-linux-gnu-gcc          - don't supported by current qemu release
# s390x-linux-gnu-gcc         - qemu troubles (hang/abort)
# sh4-linux-gnu-gcc           - qemu troubles (pread syscall, etc)
# mips64-linux-gnuabi64-gcc   - qemu troubles (pread syscall, etc)
# sparc64-linux-gnu-gcc       - qemu troubles (fcntl for F_SETLK/F_GETLK)
# alpha-linux-gnu-gcc         - qemu (or gcc) troubles (coredump)

CROSS_LIST_NOQEMU = hppa-linux-gnu-gcc s390x-linux-gnu-gcc \
	sh4-linux-gnu-gcc mips64-linux-gnuabi64-gcc \
	sparc64-linux-gnu-gcc alpha-linux-gnu-gcc

cross-gcc:
	@echo "CORRESPONDING CROSS-COMPILERs ARE REQUIRED."
	@echo "FOR INSTANCE: apt install g++-aarch64-linux-gnu g++-alpha-linux-gnu g++-arm-linux-gnueabihf g++-hppa-linux-gnu g++-mips-linux-gnu g++-mips64-linux-gnuabi64 g++-powerpc-linux-gnu g++-powerpc64-linux-gnu g++-s390x-linux-gnu g++-sh4-linux-gnu g++-sparc64-linux-gnu"
	@for CC in $(CROSS_LIST_NOQEMU) $(CROSS_LIST); do \
		echo "===================== $$CC"; \
		$(MAKE) clean && CC=$$CC CXX=$$(echo $$CC | sed 's/-gcc/-g++/') EXE_LDFLAGS=-static $(MAKE) all || exit $$?; \
	done

# Unfortunately qemu don't provide robust support for futexes.
# Therefore it is impossible to run full multi-process tests.
cross-qemu:
	@echo "CORRESPONDING CROSS-COMPILERs AND QEMUs ARE REQUIRED."
	@echo "FOR INSTANCE: "
	@echo "	1) apt install g++-aarch64-linux-gnu g++-alpha-linux-gnu g++-arm-linux-gnueabihf g++-hppa-linux-gnu g++-mips-linux-gnu g++-mips64-linux-gnuabi64 g++-powerpc-linux-gnu g++-powerpc64-linux-gnu g++-s390x-linux-gnu g++-sh4-linux-gnu g++-sparc64-linux-gnu"
	@echo "	2) apt install binfmt-support qemu-user-static qemu-user qemu-system-arm qemu-system-mips qemu-system-misc qemu-system-ppc qemu-system-sparc"
	@for CC in $(CROSS_LIST); do \
		echo "===================== $$CC + qemu"; \
		$(MAKE) clean && \
			CC=$$CC CXX=$$(echo $$CC | sed 's/-gcc/-g++/') EXE_LDFLAGS=-static XCFLAGS="-DMDBX_SAFE4QEMU $(XCFLAGS)" \
			$(MAKE) check-singleprocess || exit $$?; \
	done

#< dist-cutoff-end
################################################################################
# Benchmarking by ioarena

IOARENA ?= $(shell \
  (test -x ../ioarena/@BUILD/src/ioarena && echo ../ioarena/@BUILD/src/ioarena) || \
  (test -x ../../@BUILD/src/ioarena && echo ../../@BUILD/src/ioarena) || \
  (test -x ../../src/ioarena && echo ../../src/ioarena) || which ioarena)
NN	?= 25000000

ifneq ($(wildcard $(IOARENA)),)

.PHONY: bench clean-bench re-bench

clean-bench:
	rm -rf bench-*.txt _ioarena/*

re-bench: clean-bench bench

define bench-rule
bench-$(1)_$(2).txt: $(3) $(IOARENA) Makefile
	LD_LIBRARY_PATH="./:$$$${LD_LIBRARY_PATH}" \
		$(IOARENA) -D $(1) -B crud -m nosync -n $(2) \
		| tee $$@ | grep throughput && \
	LD_LIBRARY_PATH="./:$$$${LD_LIBRARY_PATH}" \
		$(IOARENA) -D $(1) -B get,iterate -m sync -r 4 -n $(2) \
		| tee -a $$@ | grep throughput \
	|| mv -f $$@ $$@.error

endef

$(eval $(call bench-rule,mdbx,$(NN),libmdbx.$(SO_SUFFIX)))

$(eval $(call bench-rule,sophia,$(NN)))
$(eval $(call bench-rule,leveldb,$(NN)))
$(eval $(call bench-rule,rocksdb,$(NN)))
$(eval $(call bench-rule,wiredtiger,$(NN)))
$(eval $(call bench-rule,forestdb,$(NN)))
$(eval $(call bench-rule,lmdb,$(NN)))
$(eval $(call bench-rule,nessdb,$(NN)))
$(eval $(call bench-rule,sqlite3,$(NN)))
$(eval $(call bench-rule,ejdb,$(NN)))
$(eval $(call bench-rule,vedisdb,$(NN)))
$(eval $(call bench-rule,dummy,$(NN)))

$(eval $(call bench-rule,debug,10))

bench: bench-mdbx_$(NN).txt

.PHONY: bench-debug

bench-debug: bench-debug_10.txt

bench-quartet: bench-mdbx_$(NN).txt bench-lmdb_$(NN).txt bench-rocksdb_$(NN).txt bench-wiredtiger_$(NN).txt

endif
