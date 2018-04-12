LDFLAGS += -X "github.com/pingcap/tidb-inspect-tools/pkg/utils.BuildTS=$(shell date -u '+%Y-%m-%d %I:%M:%S')"
LDFLAGS += -X "github.com/pingcap/tidb-inspect-tools/pkg/utils.GitHash=$(shell git rev-parse HEAD)"

GO=GO15VENDOREXPERIMENT="1" go
GOTEST=GO15VENDOREXPERIMENT="1" CGO_ENABLED=1 go test
PACKAGES := $$(go list ./... | grep -vE 'vendor')

.PHONY: build grafana_collector clean

build: check test grafana_collector

grafana_collector:
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/grafana_collector cmd/grafana_collector/*.go

test:
	@export log_level=error; \
	$(GOTEST) -cover $(PACKAGES)

check:
	$(GO) get github.com/golang/lint/golint

	$(GO) tool vet . 2>&1 | grep -vE 'vendor' | awk '{print} END{if(NR>0) {exit 1}}'
	$(GO) tool vet --shadow . 2>&1 | grep -vE 'vendor' | awk '{print} END{if(NR>0) {exit 1}}'
	golint ./... 2>&1 | grep -vE 'vendor' | awk '{print} END{if(NR>0) {exit 1}}'
	gofmt -s -l . 2>&1 | grep -vE 'vendor' | awk '{print} END{if(NR>0) {exit 1}}'

update:
	which dep 2>/dev/null || go get -u github.com/golang/dep/cmd/dep
ifdef PKG
	dep ensure -add ${PKG}
else
	dep ensure -update
endif
	dep prune

clean:
	@rm -rf bin/*
