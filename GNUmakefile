default: install

.PHONY: install
install:
	go install .

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

# Run example stax_accounts datasource
.PHONY: datasource-stax_accounts
datasource-stax_accounts:
	terraform -chdir=examples/data-sources/stax_accounts plan

# Run example stax_account_types datasource
.PHONY: datasource-stax_account_types
datasource-stax_account_types:
	terraform -chdir=examples/data-sources/stax_account_types plan -var="account_type_id=$(ACCOUNT_TYPE_ID)"

# Run example stax_account_types datasource
.PHONY: datasource-stax_groups
datasource-stax_groups:
	terraform -chdir=examples/data-sources/stax_groups plan -var="group_id=$(GROUP_ID)"

# Run example stax_account resource plan
.PHONY: account-resource-plan
account-resource-plan:
	terraform -chdir=examples/resources/stax_account plan -var="account_type_id=$(ACCOUNT_TYPE_ID)"

# Run example stax_account resource apply
.PHONY: account-resource-apply
account-resource-apply:
	terraform -chdir=examples/resources/stax_account apply -var="account_type_id=$(ACCOUNT_TYPE_ID)"

# Run example stax_account_type resource plan
.PHONY: account-type-resource-plan
account-type-resource-plan:
	terraform -chdir=examples/resources/stax_account_type plan

# Run example stax_account_type resource apply
.PHONY: account-type-resource-apply
account-type-resource-apply:
	terraform -chdir=examples/resources/stax_account_type apply


# Run example stax_group resource plan
.PHONY: group-resource-plan
group-resource-plan:
	terraform -chdir=examples/resources/stax_group plan

# Run example stax_group resource apply
.PHONY: group-resource-apply
group-resource-apply:
	terraform -chdir=examples/resources/stax_group apply

# Run example stax_account import
.PHONY: account-resource-import
account-resource-import:
	rm -rf examples/resources/stax_account/*.tfstate
	cd examples/resources/stax_account && terraform import -var="account_type_id=$(ACCOUNT_TYPE_ID)" stax_account.presentation-dev $(IMPORT_STAX_ACCOUNT_ID)

# Run example stax_account_type import
.PHONY: account-type-resource-import
account-type-resource-import:
	rm -rf examples/resources/stax_account_type/*.tfstate
	cd examples/resources/stax_account_type && terraform import stax_account_type.production $(IMPORT_STAX_ACCOUNT_TYPE_ID)

# Run example stax_group import
.PHONY: group-resource-import
group-resource-import:
	rm -rf examples/resources/stax_group/*.tfstate
	cd examples/resources/stax_group && terraform import stax_group.cost-data-scientist $(IMPORT_STAX_GROUP_ID)
