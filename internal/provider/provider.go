package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure StaxProvider satisfies various provider interfaces.
var _ provider.Provider = &StaxProvider{}

// StaxProvider defines the provider implementation.
type StaxProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// StaxProviderModel describes the provider data model.
type StaxProviderModel struct {
	Installation      types.String `tfsdk:"installation"`
	APITokenAccessKey types.String `tfsdk:"api_token_access_key"`
	APITokenSecretKey types.String `tfsdk:"api_token_secret_key"`
}

func (p *StaxProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "Stax"
	resp.Version = p.version
}

func (p *StaxProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"installation": schema.StringAttribute{
				MarkdownDescription: "Installation name",
				Required:            true,
			},
			"api_token_access_key": schema.StringAttribute{
				MarkdownDescription: "API Token Access Key",
				Required:            true,
			},
			"api_token_secret_key": schema.StringAttribute{
				MarkdownDescription: "API Token Secret Key",
				Required:            true,
			},
		},
	}
}

func (p *StaxProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data StaxProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources
	client := http.DefaultClient
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *StaxProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAccountResource,
	}
}

func (p *StaxProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &StaxProvider{
			version: version,
		}
	}
}
