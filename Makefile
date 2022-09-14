REPO?=apm
NAME?=webhook
TAG?=latest

GO_LICENSER_VERSION?=v0.4.0
GO_LICENSER=github.com/elastic/go-licenser@$(GO_LICENSER_VERSION)
GO_CI_LINT_VERSION?=v1.48.0
GO_CI_LINT=github.com/golangci/golangci-lint/cmd/golangci-lint@$(GO_CI_LINT_VERSION)

export HELM_INSTALL_DIR=$(PWD)/bin
HELM=$(HELM_INSTALL_DIR)/helm
HELM_CHART?=./apm-attacher
HELM_CHART_NAME?=dev-apm-attacher

.PHONY: gen-notice
gen-notice:
	@bash ./scripts/notice.sh

check-licenses:
	go run $(GO_LICENSER) -d .
	go run $(GO_LICENSER) -d -ext .java .
	go run $(GO_LICENSER) -d -ext .js .

update-licenses:
	go run $(GO_LICENSER) .
	go run $(GO_LICENSER) -ext .java .

lint:
	go run $(GO_CI_LINT) version
	go run $(GO_CI_LINT) run

.webhook: *.go Dockerfile
	docker build --build-arg GO_VERSION=$(shell cat .go-version) -t $(REPO)/$(NAME):$(TAG) .
	@touch $@

bin:
	@ mkdir -p $@

$(HELM): bin
	@curl -fsSL -o bin/get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
	@chmod u+x bin/get_helm.sh
	@env USE_SUDO=false ./bin/get_helm.sh

.PHONY: clean
clean:
	@ rm -rf .webhook bin

.PHONY: uninstall-chart
uninstall-chart: $(HELM)
	@ $(HELM) uninstall $(HELM_CHART_NAME)

.PHONY: install-chart
install-chart: $(HELM)
	@ $(HELM) upgrade $(HELM_CHART_NAME) $(HELM_CHART) --install

.PHONY: show-version
show-version:
	@echo v$(shell grep version: apm-attacher/Chart.yaml | cut -d':' -f2 | tr -d '"'  | tr -d ' ')
