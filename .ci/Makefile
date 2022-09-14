SHELL = /bin/bash -eo pipefail

.PHONY: release-notes
release-notes: validate-branch-name
	@gh release list
	@gh \
		release \
		create $(BRANCH_NAME) \
		--title '$(BRANCH_NAME)' \
		--generate-notes

.PHONY: validate-branch-name
validate-branch-name:
ifndef BRANCH_NAME
	$(error BRANCH_NAME is undefined)
endif
