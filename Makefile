REPO?=stuartnelson3
APPNAME?=webhook

.webhook: *.go Dockerfile
	docker build -t $(REPO)/$(APPNAME) .
	touch $@
