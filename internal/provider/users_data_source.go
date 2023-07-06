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

var _ datasource.DataSource = &UsersDataSource{}

func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

// UsersDataSource defines the data source implementation.
type UsersDataSource struct {
	client staxsdk.ClientInterface
}

type UserDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	FirstName  types.String `tfsdk:"first_name"`
	LastName   types.String `tfsdk:"last_name"`
	Email      types.String `tfsdk:"email"`
	Role       types.String `tfsdk:"role"`
	Status     types.String `tfsdk:"status"`
	AuthOrigin types.String `tfsdk:"auth_origin"`
	CreatedTS  types.String `tfsdk:"created_ts"`
	ModifiedTS types.String `tfsdk:"modified_ts"`
}

// UsersDataSourceModel describes the data source data model.
type UsersDataSourceModel struct {
	ID      types.String          `tfsdk:"id"`
	Filters *UsersFiltersModel    `tfsdk:"filters"`
	Users   []UserDataSourceModel `tfsdk:"users"`
}

type UsersFiltersModel struct {
	IDs types.List `tfsdk:"ids"`
}

func (d *UsersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *UsersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Users datasource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Identifier used to select an user, this takes precedence over filters",
			},
			"filters": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"ids": schema.ListAttribute{
						MarkdownDescription: "A list of identifiers used to filter stax users",
						Optional:            true,
						ElementType:         types.StringType,
					},
				},
			},
			"users": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The identifier of the stax user",
							Computed:            true,
						},
						"first_name": schema.StringAttribute{
							MarkdownDescription: "The first name of the stax user",
							Required:            true,
						},
						"last_name": schema.StringAttribute{
							MarkdownDescription: "The last name of the stax user",
							Required:            true,
						},
						"email": schema.StringAttribute{
							MarkdownDescription: "The email of the stax user",
							Computed:            true,
						},
						"role": schema.StringAttribute{
							MarkdownDescription: "The role of the stax user",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "The status of the stax user",
							Computed:            true,
						},
						"auth_origin": schema.StringAttribute{
							MarkdownDescription: "The authentication origin of the stax user",
							Computed:            true,
						},
						"created_ts": schema.StringAttribute{
							MarkdownDescription: "The created timestamp for the stax user",
							Computed:            true,
						},
						"modified_ts": schema.StringAttribute{
							MarkdownDescription: "The modified timestamp for the stax user",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *UsersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UsersDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	userIDs := make([]string, 0)

	// given that the id takes precedence over filters, if it is set ignore filters.
	if !data.ID.IsNull() {
		userIDs = []string{data.ID.ValueString()}
	} else {
		if data.Filters != nil {
			resp.Diagnostics.Append(data.Filters.IDs.ElementsAs(ctx, &userIDs, false)...)
		}
	}

	usersResp, err := d.client.UserRead(ctx, userIDs)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read groups, got error: %s", err))
		return
	}

	tflog.Info(ctx, "reading users", map[string]interface{}{
		"count": len(usersResp.JSON200.Users),
	})

	for _, user := range usersResp.JSON200.Users {

		var email *string
		if user.Email != nil {
			email = (*string)(user.Email)
		}

		data.Users = append(data.Users, UserDataSourceModel{
			ID:         types.StringValue(aws.ToString(user.Id)),
			FirstName:  types.StringPointerValue(user.FirstName),
			LastName:   types.StringPointerValue(user.LastName),
			Status:     types.StringPointerValue(userStatusToString(user.Status)),
			Role:       types.StringPointerValue(userRoleToString(user.Role)),
			Email:      types.StringPointerValue(email),
			AuthOrigin: types.StringPointerValue(user.AuthOrigin),
			CreatedTS:  types.StringPointerValue(timeToStringPtr(user.CreatedTS)),
			ModifiedTS: types.StringPointerValue(timeToStringPtr(user.ModifiedTS)),
		})
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "read users from data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func userStatusToString(u *models.UserStatus) *string {
	if u == nil {
		return nil
	}

	status := string(*u)

	return &status
}

func userRoleToString(u *models.Role) *string {
	if u == nil {
		return nil
	}

	status := string(*u)

	return &status
}
