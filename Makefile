LDFLAGS += -X "github.com/pingcap/tidb-inspect-tools/pkg/utils.Version=1.0.0"
LDFLAGS += -X "github.com/pingcap/tidb-inspect-tools/pkg/utils.BuildTS=$(shell date -u '+%Y-%m-%d %I:%M:%S')"
LDFLAGS += -X "github.com/pingcap/tidb-inspect-tools/pkg/utils.GitHash=$(shell git rev-parse HEAD)"

GO=GO15VENDOREXPERIMENT="1" go
GOTEST=GO15VENDOREXPERIMENT="1" CGO_ENABLED=1 go test
PACKAGES := $$(go list ./... | grep -vE 'vendor')

.PHONY: build tidb_exporter tikv_metrics_proxy grafana_collector kafka_adapter syslog_adapter tcp_prober clean

build: check test tidb_exporter tikv_metrics_proxy grafana_collector kafka_adapter syslog_adapter tcp_prober

tidb_exporter:
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/tidb_exporter tidb_exporter/*.go

tikv_metrics_proxy:
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/tikv_metrics_proxy tikv_metrics_proxy/*.go

grafana_collector:
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/grafana_collector cmd/grafana_collector/*.go

kafka_adapter:
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/kafka_adapter kafka_adapter/*.go

syslog_adapter:
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/syslog_adapter syslog_adapter/*.go

tcp_prober:
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/tcp_prober tcp_prober/*.go

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
