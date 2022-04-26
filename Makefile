APPNAME?=webhook

.webhook: *.go Dockerfile.webhook
	docker build -f Dockerfile.webhook -t stuartnelson3/$(APPNAME) .
	touch $@
