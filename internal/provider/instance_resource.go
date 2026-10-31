package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = (*InstanceResource)(nil)
var _ resource.ResourceWithConfigure = (*InstanceResource)(nil)

type InstanceResource struct{ client *APIClient }
type InstanceModel struct {
	ID        types.String `tfsdk:"id"`
	TenantID  types.String `tfsdk:"tenant_id"`
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	Image     types.String `tfsdk:"image"`
	Profile   types.String `tfsdk:"profile"`
	CloudInit types.String `tfsdk:"cloud_init"`
	Project   types.String `tfsdk:"project"`
}

func NewInstanceResource() resource.Resource { return &InstanceResource{} }
func (r *InstanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}
func (r *InstanceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":         schema.StringAttribute{Computed: true, Description: "Platform instance identifier."},
		"tenant_id":  schema.StringAttribute{Required: true, PlanModifiers: replace, Description: "Tenant UUID in which to create the instance."},
		"name":       schema.StringAttribute{Required: true, PlanModifiers: replace, Description: "Instance name (lowercase letters, digits, and hyphens)."},
		"type":       schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("container"), PlanModifiers: replace, Description: "Workload type: container or virtual-machine."},
		"image":      schema.StringAttribute{Required: true, PlanModifiers: replace, Description: "Supported Incus image alias, for example ubuntu/24.04."},
		"profile":    schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("micro"), PlanModifiers: replace, Description: "Sizing profile: micro, small, medium, large, or xlarge."},
		"cloud_init": schema.StringAttribute{Optional: true, Sensitive: true, PlanModifiers: replace, Description: "Optional cloud-init document for a cloud-enabled image."},
		"project":    schema.StringAttribute{Computed: true, Description: "Incus project assigned to the tenant."},
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
		ID        string `json:"id"`
		TenantID  string `json:"tenantId"`
		Name      string `json:"name"`
		Type      string `json:"type"`
		Image     string `json:"image"`
		Profile   string `json:"profile"`
		CloudInit string `json:"cloudInit"`
		Project   string `json:"project"`
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
