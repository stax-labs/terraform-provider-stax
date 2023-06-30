package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stax-labs/terraform-provider-stax/internal/api/auth"
	"github.com/stax-labs/terraform-provider-stax/internal/api/staxsdk"
)

const (
	installationEnvVar = "STAX_INSTALLATION"
	accessKeyEnvVar    = "STAX_ACCESS_KEY"
	secretKeyEnvVar    = "STAX_SECRET_KEY"
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
	Installation              types.String `tfsdk:"installation"`
	EndpointURL               types.String `tfsdk:"endpoint_url"`
	PermissionSetsEndpointURL types.String `tfsdk:"permission_sets_endpoint_url"`
	APITokenAccessKey         types.String `tfsdk:"api_token_access_key"`
	APITokenSecretKey         types.String `tfsdk:"api_token_secret_key"`
}

func (p *StaxProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "stax"
	resp.Version = p.version
}

func (p *StaxProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"installation": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("[Stax Short Installation ID](https://support.stax.io/hc/en-us/articles/4537150525071-Stax-Installation-Regions) for your Stax tenancy's control plane. Alternatively, can be configured using the `%s` environment variable. Must provide only one of `installation` or `endpoint_url`.", installationEnvVar),
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"au1", "us1", "eu1"}...),
				},
			},
			"endpoint_url": schema.StringAttribute{
				MarkdownDescription: "Stax API endpoint for your Stax tenancy's control plane, this is used for testing and customers should use `installation`. Must provide only one of `installation` or `endpoint_url`.",
				Optional:            true,
			},
			"permission_sets_endpoint_url": schema.StringAttribute{
				MarkdownDescription: "Stax Permission Sets API endpoint for your Stax tenancy's control plane, this is used for testing and customers should use `installation`. Must provide only one of `installation` or `endpoint_url`.",
				Optional:            true,
			},
			"api_token_access_key": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Stax [API Token](https://www.stax.io/developer/api-tokens/) Access Key. Alternatively, can be configured using the `%s` environment variable.", accessKeyEnvVar),
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
						"must be a valid UUID v4",
					),
				},
			},
			"api_token_secret_key": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Stax [API Token](https://www.stax.io/developer/api-tokens/) Secret Key. Alternatively, can be configured using the `%s` environment variable.", secretKeyEnvVar),
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z0-9]*$`),
						"must only contain only alphanumeric characters",
					),
				},
				Sensitive: true,
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
		client, err := staxsdk.NewClient(ctx, &auth.APIToken{}, staxsdk.WithEndpointURL(testEndpointURL), staxsdk.WithPermissionSetsEndpointURL(testEndpointURL), staxsdk.WithAuthRequestSigner(func(ctx context.Context, req *http.Request) error {
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

	installationOpt := resolveEndpointConfiguration(ctx, data, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	apiToken := resolveAPIToken(data, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := staxsdk.NewClient(
		ctx,
		apiToken,
		append(installationOpt, staxsdk.WithUserAgentVersion(fmt.Sprintf("terraform-provider-stax/%s", p.version)))...,
	)
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
		NewGroupResource,
		NewPermissionSetResource,
		NewPermissionSetAssignmentResource,
	}
}

func (p *StaxProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAccountsDataSource,
		NewAccountTypesDataSource,
		NewGroupsDataSource,
		NewPermissionSetsDataSource,
		NewPermissionSetAssignmentsDataSource,
	}
}

func resolveEndpointConfiguration(ctx context.Context, data StaxProviderModel, resp *provider.ConfigureResponse) []staxsdk.ClientOption {

	// if an endpoint is configured use it
	if !data.EndpointURL.IsNull() && !data.PermissionSetsEndpointURL.IsNull() {
		tflog.Info(ctx, "endpoint", map[string]interface{}{
			"EndpointURL":               data.EndpointURL.ValueString(),
			"PermissionSetsEndpointURL": data.PermissionSetsEndpointURL.ValueString(),
		})

		return []staxsdk.ClientOption{
			staxsdk.WithEndpointURL(data.EndpointURL.ValueString()),
			staxsdk.WithPermissionSetsEndpointURL(data.PermissionSetsEndpointURL.ValueString()),
		}
	}

	// if an installation is configured use it
	if !data.Installation.IsNull() {
		return []staxsdk.ClientOption{staxsdk.WithInstallation(data.Installation.ValueString())}
	}

	if installation := os.Getenv(installationEnvVar); installation != "" {
		return []staxsdk.ClientOption{staxsdk.WithInstallation(installation)}
	}

	resp.Diagnostics.AddAttributeError(
		path.Root("installation"),
		"Unknown Stax Installation",
		"The provider cannot create the Stax API client as there is an unknown configuration value for the Stax Installation. "+
			fmt.Sprintf("Either target apply the source of the value first, set the value statically in the configuration, or use the %s environment variable.", installationEnvVar),
	)

	return nil
}

func resolveAPIToken(data StaxProviderModel, resp *provider.ConfigureResponse) *auth.APIToken {

	apiToken := &auth.APIToken{
		AccessKey: os.Getenv(accessKeyEnvVar),
		SecretKey: os.Getenv(secretKeyEnvVar),
	}

	if !data.APITokenAccessKey.IsNull() {
		apiToken.AccessKey = data.APITokenAccessKey.ValueString()
	}

	if !data.APITokenSecretKey.IsNull() {
		apiToken.SecretKey = data.APITokenSecretKey.ValueString()
	}

	if apiToken.AccessKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token_access_key"),
			"Unknown Stax API Token Access Key",
			"The provider cannot create the Stax API client as there is an unknown configuration value for the Stax API Token Access Key. "+
				fmt.Sprintf("Either target apply the source of the value first, set the value statically in the configuration, or use the %s environment variable.", accessKeyEnvVar),
		)
	}

	if apiToken.SecretKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token_secret_key"),
			"Unknown Stax API Token Secret Key",
			"The provider cannot create the Stax API client as there is an unknown configuration value for the Stax API Token Access Key. "+
				fmt.Sprintf("Either target apply the source of the value first, set the value statically in the configuration, or use the %s environment variable.", secretKeyEnvVar),
		)
	}

	return apiToken
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &StaxProvider{
			version: version,
		}
	}
}
