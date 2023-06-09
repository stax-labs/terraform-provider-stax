package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stax-labs/terraform-provider-stax/internal/api/auth"
	"github.com/stax-labs/terraform-provider-stax/internal/api/staxsdk"
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
	resp.TypeName = "stax"
	resp.Version = p.version
}

func (p *StaxProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"installation": schema.StringAttribute{
				MarkdownDescription: "Stax Installation Region",
				Optional:            true,
			},
			"api_token_access_key": schema.StringAttribute{
				MarkdownDescription: "API Token Access Key",
				Optional:            true,
			},
			"api_token_secret_key": schema.StringAttribute{
				MarkdownDescription: "API Token Secret Key",
				Optional:            true,
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

	if testEndpointURL := os.Getenv("INTEGRATION_TEST_ENDPOINT_URL"); testEndpointURL != "" {
		client, err := staxsdk.NewClient(&auth.APIToken{}, staxsdk.WithEndpointURL(testEndpointURL), staxsdk.WithAuthRequestSigner(func(ctx context.Context, req *http.Request) error {
			return nil
		}))
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create client, got error: %s", err))
			return
		}

		resp.DataSourceData = client
		resp.ResourceData = client

		return
	}

	tflog.Trace(ctx, "connecting to stax API", map[string]interface{}{
		"installation": data.Installation.String(),
		"access_key":   data.APITokenAccessKey.String(),
	})

	apiToken := &auth.APIToken{
		AccessKey: data.APITokenAccessKey.ValueString(),
		SecretKey: data.APITokenSecretKey.ValueString(),
	}

	client, err := staxsdk.NewClient(apiToken, staxsdk.WithInstallation(data.Installation.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create client, got error: %s", err))
		return
	}

	err = client.Authenticate(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to authenticate client, got error: %s", err))
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *StaxProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAccountResource,
		NewAccountTypeResource,
	}
}

func (p *StaxProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAccountsDataSource,
		NewAccountTypesDataSource,
		NewGroupsDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &StaxProvider{
			version: version,
		}
	}
}
