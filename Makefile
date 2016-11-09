APP=gorjun
CC=go
VERSION=4.0.6

LDFLAGS=-ldflags "-w -s -X main.Version=${VERSION}"

all:
	$(CC) build ${LDFLAGS} -o $(APP)

