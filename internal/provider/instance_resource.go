package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = (*InstanceResource)(nil)
var _ resource.ResourceWithConfigure = (*InstanceResource)(nil)

type InstanceResource struct{ client *APIClient }
type InstanceModel struct {
	ID              types.String `tfsdk:"id"`
	TenantID        types.String `tfsdk:"tenant_id"`
	Name            types.String `tfsdk:"name"`
	Type            types.String `tfsdk:"type"`
	Image           types.String `tfsdk:"image"`
	ImageSecret     types.String `tfsdk:"image_secret"`
	Profile         types.String `tfsdk:"profile"`
	CloudInit       types.String `tfsdk:"cloud_init"`
	Project         types.String `tfsdk:"project"`
	Environment     types.Map    `tfsdk:"environment"`
	PersistentPaths types.Set    `tfsdk:"persistent_paths"`
}

func NewInstanceResource() resource.Resource { return &InstanceResource{} }
func (r *InstanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}
func (r *InstanceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":               schema.StringAttribute{Computed: true, Description: "Platform instance identifier."},
		"tenant_id":        schema.StringAttribute{Required: true, PlanModifiers: replace, Description: "Tenant UUID in which to create the instance."},
		"name":             schema.StringAttribute{Required: true, PlanModifiers: replace, Description: "Instance name (lowercase letters, digits, and hyphens)."},
		"type":             schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("container"), PlanModifiers: replace, Description: "Workload type: container, virtual-machine, or OCI application."},
		"image":            schema.StringAttribute{Required: true, PlanModifiers: replace, Description: "Incus image alias, or a GHCR image reference for application workloads."},
		"image_secret":     schema.StringAttribute{Optional: true, Sensitive: true, PlanModifiers: replace, Description: "Registry pull token for private GHCR application images. Stored in Terraform state."},
		"profile":          schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("micro"), PlanModifiers: replace, Description: "Sizing profile: micro, small, medium, large, or xlarge."},
		"cloud_init":       schema.StringAttribute{Optional: true, Sensitive: true, PlanModifiers: replace, Description: "Optional cloud-init document for a cloud-enabled image."},
		"environment":      schema.MapAttribute{Optional: true, Sensitive: true, ElementType: types.StringType, PlanModifiers: []planmodifier.Map{mapplanmodifier.RequiresReplace()}, Description: "Environment variables for a container or OCI application. Values are stored in Terraform state."},
		"persistent_paths": schema.SetAttribute{Optional: true, ElementType: types.StringType, PlanModifiers: []planmodifier.Set{setplanmodifier.RequiresReplace()}, Description: "Paths backed by persistent volumes on the shared NAS storage pool. Volumes survive instance replacement and deletion."},
		"project":          schema.StringAttribute{Computed: true, Description: "Incus project assigned to the tenant."},
	}}
}
func (r *InstanceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected *APIClient, got %T", req.ProviderData))
		return
	}
	r.client = c
}
func (r *InstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan InstanceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := map[string]any{"tenantId": plan.TenantID.ValueString(), "name": plan.Name.ValueString(), "type": plan.Type.ValueString(), "image": plan.Image.ValueString(), "profile": plan.Profile.ValueString()}
	if !plan.ImageSecret.IsNull() && !plan.ImageSecret.IsUnknown() {
		input["imageSecret"] = plan.ImageSecret.ValueString()
	}
	if !plan.Environment.IsNull() && !plan.Environment.IsUnknown() {
		environment := map[string]string{}
		resp.Diagnostics.Append(plan.Environment.ElementsAs(ctx, &environment, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input["environment"] = environment
	}
	if !plan.PersistentPaths.IsNull() && !plan.PersistentPaths.IsUnknown() {
		var paths []string
		resp.Diagnostics.Append(plan.PersistentPaths.ElementsAs(ctx, &paths, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input["persistentPaths"] = paths
	}
	if !plan.CloudInit.IsNull() && !plan.CloudInit.IsUnknown() {
		input["cloudInit"] = plan.CloudInit.ValueString()
	}
	var result struct {
		ID      string `json:"id"`
		Project string `json:"project"`
	}
	_, err := r.client.Request(ctx, "POST", "/instances", input, &result)
	if err != nil {
		resp.Diagnostics.AddError("Create instance failed", err.Error())
		return
	}
	plan.ID = types.StringValue(result.ID)
	plan.Project = types.StringValue(result.Project)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *InstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state InstanceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var result struct {
		ID              string            `json:"id"`
		TenantID        string            `json:"tenantId"`
		Name            string            `json:"name"`
		Type            string            `json:"type"`
		Image           string            `json:"image"`
		Profile         string            `json:"profile"`
		CloudInit       string            `json:"cloudInit"`
		Environment     map[string]string `json:"environment"`
		PersistentPaths []string          `json:"persistentPaths"`
		Project         string            `json:"project"`
	}
	status, err := r.client.Request(ctx, "GET", "/instances/"+pathID(state.ID.ValueString()), nil, &result)
	if status == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read instance failed", err.Error())
		return
	}
	state.ID = types.StringValue(result.ID)
	state.TenantID = types.StringValue(result.TenantID)
	state.Name = types.StringValue(result.Name)
	state.Type = types.StringValue(result.Type)
	state.Image = types.StringValue(result.Image)
	state.Profile = types.StringValue(result.Profile)
	if result.CloudInit == "" {
		state.CloudInit = types.StringNull()
	} else {
		state.CloudInit = types.StringValue(result.CloudInit)
	}
	state.Project = types.StringValue(result.Project)
	if len(result.Environment) == 0 {
		state.Environment = types.MapNull(types.StringType)
	} else {
		value, diagnostics := types.MapValueFrom(ctx, types.StringType, result.Environment)
		state.Environment = value
		resp.Diagnostics.Append(diagnostics...)
	}
	if len(result.PersistentPaths) == 0 {
		state.PersistentPaths = types.SetNull(types.StringType)
	} else {
		value, diagnostics := types.SetValueFrom(ctx, types.StringType, result.PersistentPaths)
		state.PersistentPaths = value
		resp.Diagnostics.Append(diagnostics...)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *InstanceResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
}
func (r *InstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state InstanceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	status, err := r.client.Request(ctx, "DELETE", "/instances/"+pathID(state.ID.ValueString()), nil, nil)
	if err != nil && status != 404 {
		resp.Diagnostics.AddError("Delete instance failed", err.Error())
	}
}
