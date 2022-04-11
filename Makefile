APPNAME?=webhook
AGENT?=agent-container

$(APPNAME).pem $(APPNAME).key:
	./gen-cert.sh $(APPNAME)

.webhook: main.go Dockerfile.webhook $(APPNAME).pem $(APPNAME).key
	docker build -f Dockerfile.webhook -t $(APPNAME) .
	touch $@

.agent: Dockerfile.agent
	docker build -f Dockerfile.agent -t $(AGENT) .
	touch $@
