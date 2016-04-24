APP=gorjun
CC=go

all:
	$(CC) build -ldflags="-w -s" -o $(APP)
