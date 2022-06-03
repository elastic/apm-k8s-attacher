REPO?=apm
NAME?=webhook
TAG?=latest

.PHONY: gen-notice
gen-notice:
	@bash ./scripts/notice.sh

.webhook: *.go Dockerfile
	docker build -t $(REPO)/$(NAME):$(TAG) .
	touch $@
