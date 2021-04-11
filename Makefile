GOBIN = $(CURDIR)/build/bin
GOBUILD = go build -trimpath -tags "mdbx"

docker_run:
	docker run -it --rm -m 64m --memory-swap 0 --memory-swappiness 0 -v $(CURDIR):/app inblocks_reproduce bash

docker: mdbx2 db-tools
	mkdir -p data_lmdb
	mkdir -p data_mdbx
	docker build -t inblocks_reproduce .
	docker run -it --rm -m 512m -v $(CURDIR):/app inblocks_reproduce bash

mdbx2:
	@echo "Building mdbx"
	@cd mdbx-go/dist/ \
	&& make clean && make config.h \
	&& echo '#define MDBX_DEBUG 0' >> config.h \
	&& echo '#define MDBX_FORCE_ASSERTIONS 0' >> config.h \
	&& echo '#define MDBX_ENABLE_MADVISE 0' >> config.h \
	&& echo '#define MDBX_TXN_CHECKOWNER 1' >> config.h \
	&& echo '#define MDBX_ENV_CHECKPID 1' >> config.h \
	&& echo '#define MDBX_DISABLE_PAGECHECKS 1' >> config.h \
	&& CFLAGS_EXTRA="-Wno-deprecated-declarations" make mdbx-static.o

db-tools: mdbx2
	mkdir -p $(GOBIN)

	@echo "Building bb-tools"
	cd lmdb/libraries/liblmdb; pwd;ls; make clean mdb_stat mdb_copy mdb_dump mdb_load; cp mdb_stat $(GOBIN); cp mdb_copy $(GOBIN); cp mdb_dump $(GOBIN); cp mdb_load $(GOBIN); cd ../../../

	cd mdbx-go/dist/ && make tools
	cp mdbx-go/dist/mdbx_chk $(GOBIN)
	cp mdbx-go/dist/mdbx_copy $(GOBIN)
	cp mdbx-go/dist/mdbx_dump $(GOBIN)
	cp mdbx-go/dist/mdbx_drop $(GOBIN)
	cp mdbx-go/dist/mdbx_load $(GOBIN)
	cp mdbx-go/dist/mdbx_stat $(GOBIN)
