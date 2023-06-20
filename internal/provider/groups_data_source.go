package provider

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stax-labs/terraform-provider-stax/internal/api/staxsdk"
)

var _ datasource.DataSource = &GroupsDataSource{}

func NewGroupsDataSource() datasource.DataSource {
	return &GroupsDataSource{}
}

// GroupsDataSource defines the data source implementation.
type GroupsDataSource struct {
	client staxsdk.ClientInterface
}

type GroupDataSourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Status types.String `tfsdk:"status"`
	Type   types.String `tfsdk:"type"`
}

// GroupsDataSourceModel describes the data source data model.
type GroupsDataSourceModel struct {
	ID      types.String           `tfsdk:"id"`
	Filters *GroupsFiltersModel    `tfsdk:"filters"`
	Groups  []GroupDataSourceModel `tfsdk:"groups"`
}

type GroupsFiltersModel struct {
	IDs types.List `tfsdk:"ids"`
}

func (d *GroupsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

func (d *GroupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Groups datasource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Group identifier used to select an group, this takes precedence over filters",
			},
			"filters": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"ids": schema.ListAttribute{
						MarkdownDescription: "A list of identifiers used to filter stax groups",
						Optional:            true,
						ElementType:         types.StringType,
					},
				},
			},
			"groups": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The identifier of the stax group",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the stax group",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "The status of the stax group",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of stax group, this can be either `LOCAL` or `SCIM`. Note that groups with a type of `SCIM` cannot be updated.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *GroupsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GroupsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	groupIDs := make([]string, 0)

	// given that the id takes precedence over filters, if it is set ignore filters.
	if !data.ID.IsNull() {
		groupIDs = []string{data.ID.ValueString()}
	} else {
		if data.Filters != nil {
			resp.Diagnostics.Append(data.Filters.IDs.ElementsAs(ctx, &groupIDs, false)...)
		}
	}

	groupsResp, err := d.client.GroupRead(ctx, groupIDs)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read groups, got error: %s", err))
		return
	}

	tflog.Info(ctx, "reading groups", map[string]interface{}{
		"count": len(groupsResp.JSON200.Groups),
	})

	for _, group := range groupsResp.JSON200.Groups {
		data.Groups = append(data.Groups, GroupDataSourceModel{
			ID:     types.StringValue(aws.ToString(group.Id)),
			Name:   types.StringValue(group.Name),
			Status: types.StringValue(string(group.Status)),
			Type:   types.StringValue(string(group.GroupType)),
		})
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "read accounts from data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
