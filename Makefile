APP=gorjun
CC=go
VERSION=4.0.0-RC10-SNAPSHOT
BUILD_TIME=`date +%FT%T%z`

LDFLAGS=-ldflags "-w -s -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

all:
	$(CC) build ${LDFLAGS} -o $(APP)

