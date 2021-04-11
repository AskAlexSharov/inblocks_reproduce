GOBIN = $(CURDIR)/build/bin
GOBUILD = go build -trimpath -tags "mdbx"

run_mdbx:
	docker run -it --rm -m 512m inblocks_reproduce

run_lmdb:
	docker run -it --rm -m 512m inblocks_reproduce lmdb

build:
	docker build -t inblocks_reproduce .

mdbx2:
	@echo "Building mdbx"
	@cd mdbx-go/dist/ \
	&& make clean && make config.h \
	&& echo '#define MDBX_DEBUG 0' >> config.h \
	&& echo '#define MDBX_FORCE_ASSERTIONS 0' >> config.h \
	&& echo '#define MDBX_ENABLE_MADVISE 0' >> config.h \
	&& echo '#define MDBX_TXN_CHECKOWNER 1' >> config.h \
	&& echo '#define MDBX_ENV_CHECKPID 1' >> config.h \
	&& echo '#define MDBX_DISABLE_PAGECHECKS 0' >> config.h \
	&& CFLAGS_EXTRA="-Wno-deprecated-declarations" make mdbx-static.o

db-tools: mdbx2
	mkdir -p $(GOBIN)

	@echo "Building bb-tools"
	go mod vendor; cd vendor/github.com/ledgerwatch/lmdb-go/dist; make clean mdb_stat mdb_copy mdb_dump mdb_drop mdb_load; cp mdb_stat $(GOBIN); cp mdb_copy $(GOBIN); cp mdb_dump $(GOBIN); cp mdb_drop $(GOBIN); cp mdb_load $(GOBIN); cd ../../../../..; rm -rf vendor
	$(GOBUILD) -o $(GOBIN)/lmdbgo_copy github.com/ledgerwatch/lmdb-go/cmd/lmdb_copy
	$(GOBUILD) -o $(GOBIN)/lmdbgo_stat github.com/ledgerwatch/lmdb-go/cmd/lmdb_stat

	cd mdbx-go/dist/ && make tools
	cp mdbx-go/dist/mdbx_chk $(GOBIN)
	cp mdbx-go/dist/mdbx_copy $(GOBIN)
	cp mdbx-go/dist/mdbx_dump $(GOBIN)
	cp mdbx-go/dist/mdbx_drop $(GOBIN)
	cp mdbx-go/dist/mdbx_load $(GOBIN)
	cp mdbx-go/dist/mdbx_stat $(GOBIN)
	cp mdbx-go/dist/mdbx_drop $(GOBIN)
