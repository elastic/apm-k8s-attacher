APPNAME?=webhook
AGENT?=agent-container

$(APPNAME).pem $(APPNAME).key:
	./gen-cert.sh $(APPNAME)

.webhook: *.go Dockerfile.webhook $(APPNAME).pem $(APPNAME).key
	docker build -f Dockerfile.webhook -t stuartnelson3/$(APPNAME) .
	touch $@

.agent: Dockerfile.agent
	docker build -f Dockerfile.agent -t stuartnelson3/$(AGENT) .
	touch $@
