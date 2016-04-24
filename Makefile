APP=gorjun
CC=go

$(APP): main.go
	$(CC) build -ldflags="-w -s" -o $@ $^

clean:
	-rm -f $(APP)
