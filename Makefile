APP=gorjun
CC=go
VERSION=$(shell git describe --abbrev=0 --tags | awk -F'.' '{print $$1"."$$2"."$$3+1}')
ifeq (${GIT_BRANCH}, )
	GIT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD | grep -iv head)
endif
ifneq (${GIT_BRANCH}, )
	#VERSION:=${VERSION}-SNAPSHOT
	VERSION:=6.0.1
endif
COMMIT=$(shell git rev-parse HEAD)

LDFLAGS=-ldflags "-w -s -X main.version=${VERSION}:${GIT_BRANCH}:${COMMIT}"

all:
	$(CC) build ${LDFLAGS} -o $(APP)

