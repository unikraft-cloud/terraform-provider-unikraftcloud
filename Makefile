.PHONY: all
.DEFAULT: all
all: help

.PHONY: testacc
testacc: ## Run acceptance tests
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

.PHONY: help
help: ## Show this help menu and exit
	@awk 'BEGIN { \
		FS = ":.*##"; \
		printf "Terraform provider targets.\n\n"; \
		printf "\033[1mUSAGE\033[0m\n"; \
		printf "  make [VAR=... [VAR=...]] \033[36mTARGET\033[0m\n\n"; \
		printf "\033[1mTARGETS\033[0m\n"; \
	} \
	/^[a-zA-Z0-9_-]+:.*?##/ { \
		printf "  \033[36m%-23s\033[0m %s\n", $$1, $$2 \
	} \
	/^##@/ { \
		printf "\n\033[1m%s\033[0m\n", substr($$0, 5) \
	} ' $(MAKEFILE_LIST)
