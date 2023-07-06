package provider

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/models"
	"github.com/stax-labs/terraform-provider-stax/internal/api/staxsdk"
)

var _ datasource.DataSource = &APITokensDataSource{}

func NewAPITokensDataSource() datasource.DataSource {
	return &APITokensDataSource{}
}

// APITokensDataSource defines the data source implementation.
type APITokensDataSource struct {
	client staxsdk.ClientInterface
}

type APITokenDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Role       types.String `tfsdk:"role"`
	Status     types.String `tfsdk:"status"`
	CreatedTS  types.String `tfsdk:"created_ts"`
	ModifiedTS types.String `tfsdk:"modified_ts"`
	Tags       types.Map    `tfsdk:"tags"`
}

// APITokensDataSourceModel describes the data source data model.
type APITokensDataSourceModel struct {
	ID        types.String              `tfsdk:"id"`
	Filters   *APITokenFiltersModel     `tfsdk:"filters"`
	APITokens []APITokenDataSourceModel `tfsdk:"api_tokens"`
}

type APITokenFiltersModel struct {
	IDs types.List `tfsdk:"ids"`
}

func (d *APITokensDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_tokens"
}

func (d *APITokensDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "API Tokens datasource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Identifier used to select an api token, this takes precedence over filters",
			},
			"filters": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"ids": schema.ListAttribute{
						MarkdownDescription: "A list of identifiers used to filter stax api tokens",
						Optional:            true,
						ElementType:         types.StringType,
					},
				},
			},
			"api_tokens": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The identifier of the stax api token",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the stax api token",
							Required:            true,
						},
						"role": schema.StringAttribute{
							MarkdownDescription: "The role of the stax api token",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "The status of the stax api token",
							Computed:            true,
						},
						"created_ts": schema.StringAttribute{
							MarkdownDescription: "The created timestamp for the stax api token",
							Computed:            true,
						},
						"modified_ts": schema.StringAttribute{
							MarkdownDescription: "The modified timestamp for the stax api token",
							Computed:            true,
						},
						"tags": schema.MapAttribute{
							MarkdownDescription: "The tags associated with the stax api token",
							Computed:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *APITokensDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*staxsdk.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *APITokensDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data APITokensDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	apiTokenIDs := make([]string, 0)

	// given that the id takes precedence over filters, if it is set ignore filters.
	if !data.ID.IsNull() {
		apiTokenIDs = []string{data.ID.ValueString()}
	} else {
		if data.Filters != nil {
			resp.Diagnostics.Append(data.Filters.IDs.ElementsAs(ctx, &apiTokenIDs, false)...)
		}
	}

	usersResp, err := d.client.APITokenRead(ctx, apiTokenIDs)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read groups, got error: %s", err))
		return
	}

	tflog.Info(ctx, "reading api tokens", map[string]interface{}{
		"count": len(usersResp.JSON200.ApiTokens),
	})

	for _, apitoken := range usersResp.JSON200.ApiTokens {

		apitokenData := APITokenDataSourceModel{
			ID:         types.StringValue(aws.ToString(apitoken.AccessKey)),
			Name:       types.StringValue(apitoken.Name),
			Status:     types.StringValue(string(apitoken.Status)),
			Role:       types.StringValue(string(apitoken.Role)),
			CreatedTS:  types.StringPointerValue(timeToStringPtr(apitoken.CreatedTS)),
			ModifiedTS: types.StringPointerValue(timeToStringPtr(apitoken.ModifiedTS)),
		}

		m := tokenTagsToMap(apitoken.Tags)

		tags, d := types.MapValueFrom(ctx, types.StringType, m)
		resp.Diagnostics.Append(d...)

		apitokenData.Tags = tags

		data.APITokens = append(data.APITokens, apitokenData)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "read api tokens from data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func tokenTagsToMap(tags *models.Tags) map[string]string {
	tokenTags := make(map[string]string)

	if tags == nil {
		return tokenTags
	}

	for k, v := range *tags {
		tokenTags[k] = v
	}

	return tokenTags
}
