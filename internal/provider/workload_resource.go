package provider

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/models"
	"github.com/stax-labs/terraform-provider-stax/internal/api/staxsdk"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &WorkloadResource{}
var _ resource.ResourceWithConfigure = &WorkloadResource{}
var _ resource.ResourceWithImportState = &WorkloadResource{}

type WorkloadResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	AccountID          types.String `tfsdk:"account_id"`
	CatalogID          types.String `tfsdk:"catalog_id"`
	CatalogueVersionID types.String `tfsdk:"catalog_version_id"`
	Region             types.String `tfsdk:"region"`
	Parameters         types.Set    `tfsdk:"parameters"`
	Tags               types.Map    `tfsdk:"tags"`
	Status             types.String `tfsdk:"status"`
}

var workloadParameterAttrTypes = map[string]attr.Type{
	"key":   types.StringType,
	"value": types.StringType,
}

type WorkloadParameterData struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

func NewWorkloadResource() resource.Resource {
	return &WorkloadResource{}
}

// WorkloadResource defines the resource implementation.
type WorkloadResource struct {
	client staxsdk.ClientInterface
}

func (r *WorkloadResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workload"
}

func (r *WorkloadResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Stax Workload resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Workload identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the workload",
				Required:            true,
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: "The identifier for the account hosting the workload",
				Required:            true,
			},
			"catalog_id": schema.StringAttribute{
				MarkdownDescription: "The workload catalog identifier for the used by the workload",
				Required:            true,
			},
			"catalog_version_id": schema.StringAttribute{
				MarkdownDescription: "The workload catalog version id for the used by the workload",
				Optional:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "The AWS region where the hosting the workload",
				Required:            true,
			},
			"tags": schema.MapAttribute{
				MarkdownDescription: "Workload tags, these are applied to any cloudformation stacks provisioned as a part of this workload",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"parameters": schema.SetAttribute{
				MarkdownDescription: "The parameters for the workload",
				ElementType:         types.ObjectType{AttrTypes: workloadParameterAttrTypes},
				Optional:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the workload",
				Computed:            true,
			},
		},
	}
}

func (r *WorkloadResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WorkloadResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *WorkloadResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	parameters, diags := expandParameter(ctx, data.Parameters)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	staxTags := make(map[string]string)
	resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &staxTags, false)...)

	createResp, err := r.client.WorkloadCreate(ctx, models.WorkloadsCreateWorkload{
		Name:               data.Name.ValueString(),
		CatalogueId:        data.CatalogID.ValueString(),
		CatalogueVersionId: data.CatalogueVersionID.ValueStringPointer(),
		AccountId:          data.AccountID.ValueString(),
		Region:             models.AwsRegion(data.Region.ValueString()),
		Tags:               (*models.Tags)(&staxTags),
		Parameters:         parameters,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create workload, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "workload create response", map[string]interface{}{
		"JSON200": createResp.JSON200,
		"body":    string(createResp.Body),
	})

	diags = r.readWorkload(ctx, aws.ToString(createResp.JSON200.WorkloadId), data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkloadResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *WorkloadResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	diags := r.readWorkload(ctx, data.ID.ValueString(), data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkloadResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *WorkloadResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	parameters, diags := expandParameter(ctx, data.Parameters)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	staxTags := make(map[string]string)
	resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &staxTags, false)...)

	updateResp, err := r.client.WorkloadUpdate(ctx, data.ID.ValueString(), models.WorkloadsUpdateWorkload{
		CatalogueId: aws.String(data.CatalogID.ValueString()),
		Parameters:  parameters,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update workload, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "workload update response", map[string]interface{}{
		"JSON200": updateResp.JSON200,
		"body":    string(updateResp.Body),
	})

	diags = r.readWorkload(ctx, data.ID.ValueString(), data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkloadResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *WorkloadResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	deleteResp, err := r.client.WorkloadDelete(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete workload, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "workload deleted", map[string]interface{}{
		"id":         data.ID.ValueString(),
		"deleteResp": deleteResp.JSON200,
	})

}

func (r *WorkloadResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *WorkloadResource) readWorkload(ctx context.Context, workloadID string, data *WorkloadResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	workloadsResp, err := r.client.WorkloadReadByID(ctx, workloadID)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read workload id=%s, got error: %s", workloadID, err))
		return diags
	}

	tflog.Info(ctx, "reading workloads", map[string]interface{}{
		"count": len(workloadsResp.JSON200.Workloads),
	})

	for _, workload := range workloadsResp.JSON200.Workloads {
		data.ID = types.StringValue(aws.ToString(workload.Id))
		data.Name = types.StringValue(workload.Name)
		data.AccountID = types.StringValue(workload.AccountId)
		data.CatalogID = types.StringValue(workload.CatalogueId)
		data.CatalogueVersionID = types.StringPointerValue(workload.CatalogueVersionId)

		if workload.Status != nil {
			data.Status = types.StringValue(string(*workload.Status))
		} else {
			data.Status = types.StringNull()
		}

		var d diag.Diagnostics

		if workload.Tags != nil {
			data.Tags, d = flattenTags(ctx, workload.Tags)
			diags.Append(d...)
		}

		if workload.Parameters != nil {
			data.Parameters, d = flattenWorkloadParameter(ctx, workload.Parameters)
			diags.Append(d...)
		}
	}

	return diags
}

func flattenTags(ctx context.Context, tags *models.Tags) (types.Map, diag.Diagnostics) {
	var diags diag.Diagnostics
	if tags == nil {
		return types.MapNull(types.StringType), diags
	}

	accountTags := make(map[string]attr.Value)
	for k, v := range *tags {
		accountTags[k] = types.StringValue(v)
	}

	return types.MapValueFrom(ctx, types.StringType, accountTags)
}

func flattenWorkloadParameter(_ context.Context, apiObject *models.Parameter) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: workloadParameterAttrTypes}

	if apiObject == nil {
		return types.SetValueMust(elemType, []attr.Value{}), diags
	}

	elems := []attr.Value{}
	for k, v := range *apiObject {

		val, ok := v.(string)
		if !ok {
			diags.AddAttributeError(path.Root("parameters"), "conversion failed for parameter", fmt.Sprintf("failed to convert parameter value to string for key: %s", k))
			continue // skip this parameter
		}

		obj := map[string]attr.Value{
			"key":   types.StringValue(k),
			"value": types.StringValue(val),
		}
		objVal, d := types.ObjectValue(workloadParameterAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}
	setVal, d := types.SetValue(elemType, elems)
	diags.Append(d...)

	return setVal, diags
}

func expandParameter(ctx context.Context, parametersSet types.Set) (*[]models.KeyValueRequestParameter, diag.Diagnostics) {
	var diags diag.Diagnostics

	if parametersSet.IsNull() {
		return nil, diags
	}

	var tfList []WorkloadParameterData
	diags.Append(parametersSet.ElementsAs(ctx, &tfList, false)...)

	var keyValueRequestParameters []models.KeyValueRequestParameter

	for _, item := range tfList {
		keyValueRequestParameters = append(keyValueRequestParameters, models.KeyValueRequestParameter{
			Key:   item.Key.ValueString(),
			Value: item.Value.ValueString(),
		})
	}

	return &keyValueRequestParameters, diags
}
