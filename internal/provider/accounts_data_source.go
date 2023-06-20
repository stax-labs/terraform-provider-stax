package provider

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/models"
	"github.com/stax-labs/terraform-provider-stax/internal/api/staxsdk"
)

var _ datasource.DataSource = &AccountsDataSource{}

func NewAccountsDataSource() datasource.DataSource {
	return &AccountsDataSource{}
}

// AccountDataSource defines the data source implementation.
type AccountsDataSource struct {
	client staxsdk.ClientInterface
}

type AccountDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	AccountTypeID   types.String `tfsdk:"account_type_id"`
	AccountType     types.String `tfsdk:"account_type"`
	Status          types.String `tfsdk:"status"`
	AWsAccountID    types.String `tfsdk:"aws_account_id"`
	AwsAccountAlias types.String `tfsdk:"aws_account_alias"`
	AWSLoginUrls    types.Object `tfsdk:"aws_login_urls"`
	Tags            types.Map    `tfsdk:"tags"`
}

var awsLoginsAttrTypes = map[string]attr.Type{
	"admin":     types.StringType,
	"developer": types.StringType,
	"readonly":  types.StringType,
}

// AccountDataSourceModel describes the data source data model.
type AccountsDataSourceModel struct {
	Filters  *AccountFiltersModel     `tfsdk:"filters"`
	Accounts []AccountDataSourceModel `tfsdk:"accounts"`
}

type AccountFiltersModel struct {
	Names types.List `tfsdk:"names"`
}

func (d *AccountsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_accounts"
}

func (d *AccountsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Accounts datasource",

		Attributes: map[string]schema.Attribute{
			"filters": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"names": schema.ListAttribute{
						MarkdownDescription: "A list of names used to filter accounts",
						Optional:            true,
						ElementType:         types.StringType,
					},
				},
			},
			"accounts": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The identifier of the stax account",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the stax account",
							Computed:            true,
						},
						"account_type_id": schema.StringAttribute{
							MarkdownDescription: "The account type identifier for this stax account",
							Computed:            true,
						},
						"account_type": schema.StringAttribute{
							MarkdownDescription: "The account type for this stax account",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "The status of the account",
							Computed:            true,
						},
						"aws_account_id": schema.StringAttribute{
							MarkdownDescription: "The aws account identifier for the stax account",
							Computed:            true,
						},
						"aws_account_alias": schema.StringAttribute{
							MarkdownDescription: "The aws account alias for the stax account",
							Computed:            true,
						},
						"tags": schema.MapAttribute{
							MarkdownDescription: "The tags associated with the Stax account",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"aws_login_urls": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"admin": schema.StringAttribute{
									MarkdownDescription: "URL for administrator access via IDAM",
									Computed:            true,
								},
								"developer": schema.StringAttribute{
									MarkdownDescription: "URL for developer access via IDAM",
									Computed:            true,
								},
								"readonly": schema.StringAttribute{
									MarkdownDescription: "URL for readonly access via IDAM",
									Computed:            true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *AccountsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AccountsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AccountsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	accountNames := new([]string)
	if data.Filters != nil {
		resp.Diagnostics.Append(data.Filters.Names.ElementsAs(ctx, accountNames, false)...)
	}

	accountsResp, err := d.client.AccountRead(ctx, nil, *accountNames)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read accounts, got error: %s", err))
		return
	}

	tflog.Info(ctx, "reading accounts", map[string]interface{}{
		"count": len(accountsResp.JSON200.Accounts),
	})

	accountTypesResp, err := d.client.AccountTypeRead(ctx, []string{})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read account types, got error: %s", err))
		return
	}

	accountTypesMap := make(map[string]string)

	for _, accountType := range accountTypesResp.JSON200.AccountTypes {
		accountTypesMap[accountType.Name] = aws.ToString(accountType.Id)
	}

	for _, account := range accountsResp.JSON200.Accounts {
		accountModel := AccountDataSourceModel{
			ID:           types.StringValue(aws.ToString(account.Id)),
			Name:         types.StringValue(account.Name),
			Status:       types.StringValue(string(*account.Status)),
			AWsAccountID: types.StringValue(aws.ToString(account.AwsAccountId)),
		}

		if account.AwsAccountAlias != nil {
			accountModel.AwsAccountAlias = types.StringValue(aws.ToString(account.AwsAccountAlias))
		}

		if account.AWSLoginURLs != nil {
			accountModel.AWSLoginUrls = types.ObjectValueMust(awsLoginsAttrTypes, map[string]attr.Value{
				"admin":     types.StringValue(aws.ToString(account.AWSLoginURLs.Admin)),
				"developer": types.StringValue(aws.ToString(account.AWSLoginURLs.Developer)),
				"readonly":  types.StringValue(aws.ToString(account.AWSLoginURLs.Readonly)),
			})
		}

		if account.AccountType != nil {
			if accountTypeID, ok := accountTypesMap[aws.ToString(account.AccountType)]; ok {
				accountModel.AccountTypeID = types.StringValue(accountTypeID)
				accountModel.AccountType = types.StringValue(aws.ToString(account.AccountType))
			}
		}

		tags := staxTagsToMap(account.Tags)

		mapVal, diag := types.MapValueFrom(ctx, types.StringType, tags)
		resp.Diagnostics.Append(diag...)

		accountModel.Tags = mapVal

		data.Accounts = append(data.Accounts, accountModel)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "read accounts from data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func staxTagsToMap(tags *models.StaxTags) map[string]string {
	accountTags := make(map[string]string)

	if tags == nil {
		return accountTags
	}

	for k, v := range *tags {
		accountTags[k] = v
	}

	return accountTags
}
