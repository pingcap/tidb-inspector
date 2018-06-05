GO=GO15VENDOREXPERIMENT="1" CGO_ENABLED=0 go
GOTEST=GO15VENDOREXPERIMENT="1" CGO_ENABLED=1 go test # go race detector requires cgo

PACKAGES := $$(go list ./...| grep -vE 'vendor')

GOFILTER := grep -vE 'vendor|render.Delims|bindata_assetfs|testutil'
GOCHECKER := $(GOFILTER) | awk '{ print } END { if (NR > 0) { exit 1 } }'

LDFLAGS += -X "github.com/pingcap/tidb-inspector-tools/util.BuildTS=$(shell date -u '+%Y-%m-%d %I:%M:%S')"
LDFLAGS += -X "github.com/pingcap/tidb-inspector-tools/util.GitHash=$(shell git rev-parse HEAD)"

GOBUILD=$(GO) build -ldflags '$(LDFLAGS)'

default: tidb_exporter

all: check test tidb_exporter

tidb_exporter:
	$(GOBUILD) -o bin/tidb_exporter *.go

test:
	$(GOTEST) --race $(PACKAGES)

check:
	go get github.com/golang/lint/golint

	@echo "vet"
	@ go tool vet . 2>&1 | $(GOCHECKER)
	@ go tool vet --shadow . 2>&1 | $(GOCHECKER)
	@echo "golint"
	@ golint ./... 2>&1 | $(GOCHECKER)
	@echo "gofmt"
	@ gofmt -s -l . 2>&1 | $(GOCHECKER)

clean:
	@rm -rf bin/*

.PHONY: all test clean tidb_exporter
