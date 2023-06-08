//go:build tools

package tools

import (
	// Documentation generation
	_ "github.com/deepmap/oapi-codegen/cmd/oapi-codegen"
	// Openapi Code generation
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
	// Mocks
	_ "github.com/vektra/mockery/v2"
)
