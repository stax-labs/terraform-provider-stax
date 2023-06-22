package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/permissionssets/models"
	"github.com/stax-labs/terraform-provider-stax/internal/api/staxsdk"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &PermissionSetResource{}
var _ resource.ResourceWithConfigure = &PermissionSetResource{}
var _ resource.ResourceWithImportState = &PermissionSetResource{}

type PermissionSetResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Status               types.String `tfsdk:"status"`
	Description          types.String `tfsdk:"description"`
	CreatedBy            types.String `tfsdk:"created_by"`
	CreatedTS            types.String `tfsdk:"created_ts"`
	MaxSessionDuration   types.Int64  `tfsdk:"max_session_duration"`
	Tags                 types.Map    `tfsdk:"tags"`
	InlinePolicies       types.Set    `tfsdk:"inline_policies"`
	AWSManagedPolicyArns types.List   `tfsdk:"aws_managed_policy_arns"`
}

var permissionSetInlinePolicyAttrTypes = map[string]attr.Type{
	"name":   types.StringType,
	"policy": types.StringType,
}

type PermissionSetInlinePolicyData struct {
	Name   types.String `tfsdk:"name"`
	Policy types.String `tfsdk:"policy"`
}

func NewPermissionSetResource() resource.Resource {
	return &PermissionSetResource{}
}

// PermissionSetResource defines the resource implementation.
type PermissionSetResource struct {
	client staxsdk.ClientInterface
}

func (r *PermissionSetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permission_set"
}

func (r *PermissionSetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Stax Permission Set resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Permission Set identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the stax Permission Set",
				Required:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the stax Permission Set, can be ACTIVE or DELETED",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the stax Permission Set",
				Optional:            true,
			},
			"created_by": schema.StringAttribute{
				MarkdownDescription: "The identifier of the stax user who created the Permission Set",
				Computed:            true,
			},
			"created_ts": schema.StringAttribute{
				MarkdownDescription: "The Permission Set was creation timestamp",
				Computed:            true,
			},
			"max_session_duration": schema.Int64Attribute{
				MarkdownDescription: "The max session duration used by this Permission Set",
				Optional:            true,
			},
			"tags": schema.MapAttribute{
				MarkdownDescription: "Permission Set tags",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"inline_policies": schema.SetAttribute{
				MarkdownDescription: "The inline policies assigned to the Permission Set",
				ElementType:         types.ObjectType{AttrTypes: permissionSetInlinePolicyAttrTypes},
				Optional:            true,
			},
			"aws_managed_policy_arns": schema.ListAttribute{
				MarkdownDescription: "A list of aws managed policy arns assigned to the Permission Set",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *PermissionSetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client

}

func (r *PermissionSetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *PermissionSetResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	inlinePolicies, diags := expandInlinePolicies(ctx, data.InlinePolicies)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	awsManagedPolicyArns, diags := expandAWSManagedPolicies(ctx, data.AWSManagedPolicyArns)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := models.CreatePermissionSetRecord{
		Name:               data.Name.ValueString(),
		Description:        data.Description.ValueStringPointer(),
		MaxSessionDuration: convertToIPtr(data.MaxSessionDuration.ValueInt64Pointer()),
		InlinePolicies:     inlinePolicies,
		AWSManagedPolicies: awsManagedPolicyArns,
	}

	created, err := r.client.PermissionSetsCreate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create permission set, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(permissionSetAPIToTFResource(ctx, *created.JSON201, data)...)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PermissionSetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *PermissionSetResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	read, err := r.client.PermissionSetsReadByID(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read permission set, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(permissionSetAPIToTFResource(ctx, *read.JSON200, data)...)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PermissionSetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *PermissionSetResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	inlinePolicies, diags := expandInlinePolicies(ctx, data.InlinePolicies)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	awsManagedPolicyArns, diags := expandAWSManagedPolicies(ctx, data.AWSManagedPolicyArns)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := models.UpdatePermissionSetRecord{
		InlinePolicies:     inlinePolicies,
		Description:        data.Description.ValueStringPointer(),
		MaxSessionDuration: convertToIPtr(data.MaxSessionDuration.ValueInt64Pointer()),
		AWSManagedPolicies: awsManagedPolicyArns,
	}

	updated, err := r.client.PermissionSetsUpdate(ctx, data.ID.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create permission set, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(permissionSetAPIToTFResource(ctx, *updated.JSON200, data)...)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PermissionSetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *PermissionSetResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	_, err := r.client.PermissionSetsDelete(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete permission set, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "permission set deleted", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

}

func (r *PermissionSetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func permissionSetAPIToTFResource(ctx context.Context, permissionSet models.PermissionSetRecord, data *PermissionSetResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	tags, d := types.MapValueFrom(ctx, types.StringType, permissionSet.Tags)
	diags.Append(d...)

	inlinePolicies, d := flattenInlinePolicies(ctx, permissionSet.InlinePolicies)
	diags.Append(d...)

	awsManagedPolicyArns, d := flattenAWSManagedPolicy(ctx, permissionSet.AWSManagedPolicies)
	diags.Append(d...)

	data.ID = types.StringValue(permissionSet.Id.String())
	data.Name = types.StringValue(permissionSet.Name)
	data.Status = types.StringValue(string(permissionSet.Status))
	data.Description = types.StringPointerValue(permissionSet.Description)
	data.CreatedBy = types.StringValue(permissionSet.CreatedBy)
	data.CreatedTS = types.StringValue(permissionSet.CreatedTS.Format(time.RFC3339))
	data.MaxSessionDuration = types.Int64PointerValue(convertToI64Ptr(permissionSet.MaxSessionDuration))
	data.Tags = tags
	data.InlinePolicies = inlinePolicies
	data.AWSManagedPolicyArns = awsManagedPolicyArns

	return diags
}

func expandInlinePolicies(ctx context.Context, parametersSet types.Set) (*[]models.PermissionSetInlinePolicy, diag.Diagnostics) {
	var diags diag.Diagnostics

	if parametersSet.IsNull() {
		return nil, diags
	}

	var tfList []PermissionSetInlinePolicyData
	diags.Append(parametersSet.ElementsAs(ctx, &tfList, false)...)

	var keyValueRequestParameters []models.PermissionSetInlinePolicy

	for _, item := range tfList {
		keyValueRequestParameters = append(keyValueRequestParameters, models.PermissionSetInlinePolicy{
			Name:   item.Name.ValueString(),
			Policy: item.Policy.ValueString(),
		})
	}

	return &keyValueRequestParameters, diags
}

func expandAWSManagedPolicies(ctx context.Context, parametersSet types.List) (*[]models.PermissionSetAWSManagedPolicy, diag.Diagnostics) {
	var diags diag.Diagnostics

	if parametersSet.IsNull() {
		return nil, diags
	}

	vals := []string{}
	diags.Append(parametersSet.ElementsAs(ctx, &vals, false)...)

	var permissionSetAWSManagedPolicies []models.PermissionSetAWSManagedPolicy
	for _, v := range vals {
		permissionSetAWSManagedPolicies = append(permissionSetAWSManagedPolicies, models.PermissionSetAWSManagedPolicy{
			PolicyArn: v,
		})
	}

	return &permissionSetAWSManagedPolicies, diags
}

func flattenInlinePolicies(_ context.Context, apiObject *[]models.PermissionSetInlinePolicy) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: permissionSetInlinePolicyAttrTypes}

	if apiObject == nil {
		return types.SetNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, v := range *apiObject {

		obj := map[string]attr.Value{
			"name":   types.StringValue(v.Name),
			"policy": types.StringValue(v.Policy),
		}
		objVal, d := types.ObjectValue(permissionSetInlinePolicyAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}
	setVal, d := types.SetValue(elemType, elems)
	diags.Append(d...)

	return setVal, diags
}

func flattenAWSManagedPolicy(ctx context.Context, apiObject *[]models.PermissionSetAWSManagedPolicy) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiObject == nil {
		return types.ListNull(types.StringType), diags
	}

	policyARNs := []string{}
	for _, v := range *apiObject {
		policyARNs = append(policyARNs, v.PolicyArn)
	}

	return types.ListValueFrom(ctx, types.StringType, policyARNs)
}
