REPO?=stuartnelson3
APPNAME?=webhook

.webhook: *.go Dockerfile
	docker build -f Dockerfile -t $(REPO)/$(APPNAME) .
	touch $@
