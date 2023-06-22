package mocks

//go:generate go run github.com/vektra/mockery/v2 --srcpkg github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/server --name ServerInterface --output .
//go:generate go run github.com/vektra/mockery/v2 --srcpkg github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/client --name ClientWithResponsesInterface --output .
