REPO?=apm
NAME?=webhook
TAG?=latest

.webhook: *.go Dockerfile
	docker build -t $(REPO)/$(NAME):$(TAG) .
	touch $@
