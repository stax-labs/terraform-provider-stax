package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stax-labs/terraform-provider-stax/internal/api/staxsdk"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &AccountTypesDataSource{}

func NewAccountTypesDataSource() datasource.DataSource {
	return &AccountTypesDataSource{}
}

// AccountTypesDataSource defines the data source implementation.
type AccountTypesDataSource struct {
	client staxsdk.ClientInterface
}

type AccountTypeDataSourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Status types.String `tfsdk:"status"`
}

type AccountTypeFiltersModel struct {
	IDs types.List `tfsdk:"ids"`
}

// AccountTypesDataSourceModel describes the data source data model.
type AccountTypesDataSourceModel struct {
	Filters      *AccountTypeFiltersModel     `tfsdk:"filters"`
	AccountTypes []AccountTypeDataSourceModel `tfsdk:"account_types"`
}

func (d *AccountTypesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account_types"
}

func (d *AccountTypesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Account Types datasource",

		Attributes: map[string]schema.Attribute{
			"filters": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"ids": schema.ListAttribute{
						MarkdownDescription: "A list of identifiers used to filter account types",
						Optional:            true,
						ElementType:         types.StringType,
					},
				},
			},
			"account_types": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The identifier of the stax account type",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the stax account type",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "The status of the stax account type",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *AccountTypesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*staxsdk.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *AccountTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AccountTypesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	accountTypeIDs := new([]string)
	if data.Filters != nil {
		resp.Diagnostics.Append(data.Filters.IDs.ElementsAs(ctx, accountTypeIDs, false)...)
	}

	accountTypesResp, err := d.client.AccountTypeRead(ctx, *accountTypeIDs)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read accounts, got error: %s", err))
		return
	}

	tflog.Info(ctx, "reading account types", map[string]interface{}{
		"count": len(accountTypesResp.JSON200.AccountTypes),
	})

	for _, accountType := range accountTypesResp.JSON200.AccountTypes {
		accountTypeModel := AccountTypeDataSourceModel{
			ID:     types.StringValue(toString(accountType.Id)),
			Name:   types.StringValue(accountType.Name),
			Status: types.StringValue(fmt.Sprintf("%s", accountType.Status)),
		}

		data.AccountTypes = append(data.AccountTypes, accountTypeModel)
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
