default: install

.PHONY: install
install:
	go install .

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

# Run example stax_accounts datasource
.PHONY: accounts-datasource-stax_accounts
accounts-datasource-stax_accounts:
	terraform -chdir=examples/data-sources/stax_accounts plan -var="installation=$(STAX_INSTALLATION)" -var="api_token_access_key=$(STAX_ACCESS_KEY)" -var="api_token_secret_key=$(STAX_SECRET_KEY)"

# Run example stax_account_types datasource
.PHONY: accounts-datasource-stax_account_types
accounts-datasource-stax_account_types:
	terraform -chdir=examples/data-sources/stax_account_types plan -var="installation=$(STAX_INSTALLATION)" -var="api_token_access_key=$(STAX_ACCESS_KEY)" -var="api_token_secret_key=$(STAX_SECRET_KEY)"

# Run example stax_account resource plan
.PHONY: account-resource-plan
account-resource-plan:
	terraform -chdir=examples/resources/stax_account plan -var="installation=$(STAX_INSTALLATION)" -var="api_token_access_key=$(STAX_ACCESS_KEY)" -var="api_token_secret_key=$(STAX_SECRET_KEY)" -var="account_type_id=$(ACCOUNT_TYPE_ID)"

# Run example stax_account resource apply
.PHONY: account-resource-apply
account-resource-apply:
	terraform -chdir=examples/resources/stax_account apply -var="installation=$(STAX_INSTALLATION)" -var="api_token_access_key=$(STAX_ACCESS_KEY)" -var="api_token_secret_key=$(STAX_SECRET_KEY)" -var="account_type_id=$(ACCOUNT_TYPE_ID)"

# Run example stax_account_type resource plan
.PHONY: account-type-resource-plan
account-type-resource-plan:
	terraform -chdir=examples/resources/stax_account_type plan -var="installation=$(STAX_INSTALLATION)" -var="api_token_access_key=$(STAX_ACCESS_KEY)" -var="api_token_secret_key=$(STAX_SECRET_KEY)"

# Run example stax_account_type resource apply
.PHONY: account-type-resource-apply
account-type-resource-apply:
	terraform -chdir=examples/resources/stax_account_type apply -var="installation=$(STAX_INSTALLATION)" -var="api_token_access_key=$(STAX_ACCESS_KEY)" -var="api_token_secret_key=$(STAX_SECRET_KEY)"

# Run example stax_account import
.PHONY: account-resource-import
account-resource-import:
	rm -rf examples/resources/stax_account/*.tfstate
	cd examples/resources/stax_account && terraform import -var="installation=$(STAX_INSTALLATION)" -var="api_token_access_key=$(STAX_ACCESS_KEY)" -var="api_token_secret_key=$(STAX_SECRET_KEY)" stax_account.presentation-dev $(IMPORT_STAX_ACCOUNT_ID)

# Run example stax_account import
.PHONY: account-type-resource-import
account-type-resource-import:
	rm -rf examples/resources/stax_account_type/*.tfstate
	cd examples/resources/stax_account_type && terraform import -var="installation=$(STAX_INSTALLATION)" -var="api_token_access_key=$(STAX_ACCESS_KEY)" -var="api_token_secret_key=$(STAX_SECRET_KEY)" stax_account_type.production $(IMPORT_STAX_ACCOUNT_TYPE_ID)
