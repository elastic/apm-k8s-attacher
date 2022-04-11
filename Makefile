BIN?=webhook
APPNAME?=webhook

$(BIN): main.go
	docker build -f Dockerfile.webhook -t $(APPNAME) .
