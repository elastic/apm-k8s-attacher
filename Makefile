APPNAME?=webhook
AGENT?=agent-container

.webhook: main.go Dockerfile.webhook
	docker build -f Dockerfile.webhook -t $(APPNAME) .
	touch $@

.agent: Dockerfile.agent
	docker build -f Dockerfile.agent -t $(AGENT) .
	touch $@
