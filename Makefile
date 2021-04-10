test: mdbx
	go build .


mdbx:
	@echo "Building mdbx"
	@cd mdbx/dist/ \
	&& make clean && make config.h \
	&& echo '#define MDBX_DEBUG 0' >> config.h \
	&& echo '#define MDBX_FORCE_ASSERTIONS 0' >> config.h \
	&& echo '#define MDBX_ENABLE_MADVISE 0' >> config.h \
	&& echo '#define MDBX_TXN_CHECKOWNER 1' >> config.h \
	&& echo '#define MDBX_ENV_CHECKPID 1' >> config.h \
	&& echo '#define MDBX_DISABLE_PAGECHECKS 0' >> config.h \
	&& CFLAGS_EXTRA="-Wno-deprecated-declarations" make mdbx-static.o

