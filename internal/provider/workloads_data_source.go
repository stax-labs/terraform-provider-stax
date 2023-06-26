package provider

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stax-labs/terraform-provider-stax/internal/api/helpers"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/models"
	"github.com/stax-labs/terraform-provider-stax/internal/api/staxsdk"
)

var _ datasource.DataSource = &WorkloadsDataSource{}

func NewWorkloadsDataSource() datasource.DataSource {
	return &WorkloadsDataSource{}
}

// WorkloadsDataSource defines the data source implementation.
type WorkloadsDataSource struct {
	client staxsdk.ClientInterface
}

type WorkloadDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	AccountID types.String `tfsdk:"account_id"`
	CatalogID types.String `tfsdk:"catalog_id"`
	Region    types.String `tfsdk:"region"`
	Status    types.String `tfsdk:"status"`
}

// WorkloadsDataSourceModel describes the data source data model.
type WorkloadsDataSourceModel struct {
	ID        types.String              `tfsdk:"id"`
	Filters   *WorkloadsFiltersModel    `tfsdk:"filters"`
	Workloads []WorkloadDataSourceModel `tfsdk:"workloads"`
}

type WorkloadsFiltersModel struct {
	IDs types.List `tfsdk:"ids"`
}

func (d *WorkloadsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workloads"
}

func (d *WorkloadsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Workloads datasource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Identifier used to select a workload, this takes precedence over filters",
			},
			"filters": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"ids": schema.ListAttribute{
						MarkdownDescription: "A list of identifiers used to filter stax workloads",
						Optional:            true,
						ElementType:         types.StringType,
					},
				},
			},
			"workloads": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The identifier of the stax workload",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the stax workload",
							Computed:            true,
						},
						"account_id": schema.StringAttribute{
							MarkdownDescription: "The account identifier for the account hosting the workload",
							Computed:            true,
						},
						"catalog_id": schema.StringAttribute{
							MarkdownDescription: "The workload catalog identifier for the used by the workload",
							Computed:            true,
						},
						"region": schema.StringAttribute{
							MarkdownDescription: "The AWS region where the workload is deployed",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "The status of the stax workload",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *WorkloadsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *WorkloadsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data WorkloadsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	workloadIDs := make([]string, 0)

	// given that the id takes precedence over filters, if it is set ignore filters.
	if !data.ID.IsNull() {
		workloadIDs = []string{data.ID.ValueString()}
	} else {
		if data.Filters != nil {
			resp.Diagnostics.Append(data.Filters.IDs.ElementsAs(ctx, &workloadIDs, false)...)
		}
	}

	workloadsResp, err := d.client.WorkloadRead(ctx, &models.WorkloadsReadWorkloadsParams{IdFilter: helpers.CommaDelimitedOptionalValue(workloadIDs)})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read workloads, got error: %s", err))
		return
	}

	tflog.Info(ctx, "reading workloads", map[string]interface{}{
		"count": len(workloadsResp.JSON200.Workloads),
	})

	for _, workload := range workloadsResp.JSON200.Workloads {
		data.Workloads = append(data.Workloads, WorkloadDataSourceModel{
			ID:        types.StringValue(aws.ToString(workload.Id)),
			Name:      types.StringValue(workload.Name),
			AccountID: types.StringValue(workload.AccountId),
			CatalogID: types.StringValue(workload.CatalogueId),
			Region:    types.StringValue(workload.Region),
			Status:    types.StringPointerValue((*string)(workload.Status)),
		})
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "read workloads from data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
