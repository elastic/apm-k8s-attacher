REPO?=apm
NAME?=webhook
TAG?=latest

.PHONY: gen-notice
gen-notice:
	@bash ./scripts/notice.sh

check-licenses:
	go install github.com/elastic/go-licenser@v0.4.0
	go run github.com/elastic/go-licenser@v0.4.0 -d .
	go run github.com/elastic/go-licenser@v0.4.0 -d -ext .java .
	go run github.com/elastic/go-licenser@v0.4.0 -d -ext .js .

update-licenses:
	go install github.com/elastic/go-licenser@v0.4.0
	go run github.com/elastic/go-licenser@v0.4.0 .
	go run github.com/elastic/go-licenser@v0.4.0 -ext .java .

.webhook: *.go Dockerfile
	docker build -t $(REPO)/$(NAME):$(TAG) .
	touch $@
