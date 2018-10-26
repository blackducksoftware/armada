CMDS = $(shell ls cmd)

ifndef REGISTRY
REGISTRY=gcr.io/gke-verification
endif

ifdef IMAGE_PREFIX
PREFIX="$(IMAGE_PREFIX)-"
endif

ifneq (, $(findstring gcr.io,$(REGISTRY))) 
PREFIX_CMD="gcloud"
DOCKER_OPTS="--"
endif

OUTDIR=_output
LOCAL_TARGET=local

CURRENT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

.PHONY: all clean test push test ${CMDS} container local

all: build

build: ${OUTDIR} $(CMDS)

${LOCAL_TARGET}: ${OUTDIR} $(CMDS)

$(CMDS):
ifeq ($(MAKECMDGOALS),${LOCAL_TARGET})
	cd cmd/$@; CGO_ENABLED=0 go build -o $@
else
	docker run --rm -e CGO_ENABLED=0 -e GOOS=linux -e GOARCH=amd64 -v "${CURRENT_DIR}":/go/src/github.com/blackducksoftware/armada -w /go/src/github.com/blackducksoftware/armada/cmd/$@ golang:1.9 go build -o $@
endif
	cp cmd/$@/$@ ${OUTDIR}

container: $(CMDS)
	$(foreach p,${CMDS},cd ${CURRENT_DIR}/cmd/$p; docker build -t $(REGISTRY)/$(PREFIX)${p} .;)

push: container
	$(foreach p,${CMDS},$(PREFIX_CMD) docker $(DOCKER_OPTS) push $(REGISTRY)/$(PREFIX)${p}:latest;)

test:
	docker run --rm -e CGO_ENABLED=0 -e GOOS=linux -e GOARCH=amd64 -v "${CURRENT_DIR}":/go/src/github.com/blackducksoftware/armada -w /go/src/github.com/blackducksoftware/armada golang:1.9 go test ./pkg/...

clean:
	rm -rf ${OUTDIR}
	$(foreach p,${CMDS},rm -f cmd/$p/$p;)

${OUTDIR}:
	mkdir -p ${OUTDIR}

lint:
	./hack/verify-gofmt.sh
	./hack/verify-golint.sh
	./hack/verify-govet.sh
