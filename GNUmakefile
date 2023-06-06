default: install

.PHONY: install
install:
	go install .

.PHONY: accounts-datasource-stax_accounts
accounts-datasource-stax_accounts:
	terraform -chdir=examples/data-sources/stax_accounts plan -var="installation=$(STAX_INSTALLATION)" -var="api_token_access_key=$(STAX_ACCESS_KEY)" -var="api_token_secret_key=$(STAX_SECRET_KEY)"

.PHONY: account-resource-plan
account-resource-plan:
	terraform -chdir=examples/resources/stax_account plan -var="installation=$(STAX_INSTALLATION)" -var="api_token_access_key=$(STAX_ACCESS_KEY)" -var="api_token_secret_key=$(STAX_SECRET_KEY)" -var="account_type_id=$(ACCOUNT_TYPE_ID)"

.PHONY: account-resource-apply
account-resource-apply:
	terraform -chdir=examples/resources/stax_account apply -var="installation=$(STAX_INSTALLATION)" -var="api_token_access_key=$(STAX_ACCESS_KEY)" -var="api_token_secret_key=$(STAX_SECRET_KEY)" -var="account_type_id=$(ACCOUNT_TYPE_ID)"

.PHONY: account-resource-import
account-resource-import:
	rm -rf examples/resources/stax_account/*.tfstate
	cd examples/resources/stax_account && terraform import -var="installation=$(STAX_INSTALLATION)" -var="api_token_access_key=$(STAX_ACCESS_KEY)" -var="api_token_secret_key=$(STAX_SECRET_KEY)" stax_account.presentation-dev $(IMPORT_STAX_ACCOUNT_ID)

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m
