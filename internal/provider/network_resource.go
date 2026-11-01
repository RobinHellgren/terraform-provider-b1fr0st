package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = (*NetworkResource)(nil)
var _ resource.ResourceWithConfigure = (*NetworkResource)(nil)
var _ resource.ResourceWithImportState = (*NetworkResource)(nil)

type NetworkResource struct{ client *APIClient }
type NetworkModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Name        types.String `tfsdk:"name"`
	Config      types.Map    `tfsdk:"config"`
	Project     types.String `tfsdk:"project"`
	Managed     types.Bool   `tfsdk:"managed"`
	Type        types.String `tfsdk:"type"`
	NetworkType types.String `tfsdk:"network_type"`
}

func NewNetworkResource() resource.Resource { return &NetworkResource{} }
func (r *NetworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}
func (r *NetworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":           schema.StringAttribute{Computed: true, Description: "Managed Incus network name."},
		"project_id":   schema.StringAttribute{Required: true, PlanModifiers: replace, Description: "Project UUID that owns this network."},
		"name":         schema.StringAttribute{Required: true, PlanModifiers: replace, Description: "Unique managed network name."},
		"network_type": schema.StringAttribute{Computed: true, Description: "Network implementation reported by the platform. Newly created managed networks are VXLAN-backed OVN overlays."},
		"config":       schema.MapAttribute{Computed: true, ElementType: types.StringType, Description: "Incus-managed network configuration. Edit supported settings in the b1fr0st console."},
		"project":      schema.StringAttribute{Computed: true, Description: "Incus project name associated with this network."},
		"managed":      schema.BoolAttribute{Computed: true, Description: "Whether Incus manages this network."},
		"type":         schema.StringAttribute{Computed: true, Description: "Incus network type."},
	}}
}
func (r *NetworkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *NetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *NetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NetworkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out struct {
		Network map[string]any `json:"network"`
	}
	_, err := r.client.Request(ctx, "POST", "/networks", map[string]any{"projectId": plan.ProjectID.ValueString(), "name": plan.Name.ValueString()}, &out)
	if err != nil {
		resp.Diagnostics.AddError("Create network failed", err.Error())
		return
	}
	resp.Diagnostics.Append(networkModelFromAPI(ctx, &plan, out.Network)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *NetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NetworkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out struct {
		Network map[string]any `json:"network"`
	}
	name := state.Name.ValueString()
	if name == "" {
		name = state.ID.ValueString()
	}
	status, err := r.client.Request(ctx, "GET", "/networks/"+pathID(name), nil, &out)
	if status == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read network failed", err.Error())
		return
	}
	resp.Diagnostics.Append(networkModelFromAPI(ctx, &state, out.Network)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *NetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NetworkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	config := map[string]string{}
	resp.Diagnostics.Append(plan.Config.ElementsAs(ctx, &config, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out struct {
		Network map[string]any `json:"network"`
	}
	_, err := r.client.Request(ctx, "PUT", "/networks/"+pathID(plan.Name.ValueString()), map[string]any{"config": config}, &out)
	if err != nil {
		resp.Diagnostics.AddError("Update network failed", err.Error())
		return
	}
	resp.Diagnostics.Append(networkModelFromAPI(ctx, &plan, out.Network)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *NetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NetworkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	status, err := r.client.Request(ctx, "DELETE", "/networks/"+pathID(state.Name.ValueString()), nil, nil)
	if err != nil && status != 404 {
		resp.Diagnostics.AddError("Delete network failed", err.Error())
	}
}
func networkModelFromAPI(ctx context.Context, model *NetworkModel, raw map[string]any) (diagnostics diag.Diagnostics) {
	if name := anyString(raw["name"]); name != "" {
		model.ID = types.StringValue(name)
		model.Name = types.StringValue(name)
	}
	if projectID := anyString(raw["projectId"]); projectID != "" {
		model.ProjectID = types.StringValue(projectID)
	}
	project := anyString(raw["tenantProject"])
	if project == "" {
		project = anyString(raw["projectName"])
	}
	if project == "" {
		project = anyString(raw["project"])
	}
	if project != "" {
		model.Project = types.StringValue(project)
	}
	if value, ok := raw["managed"].(bool); ok {
		model.Managed = types.BoolValue(value)
	}
	if networkType := anyString(raw["type"]); networkType != "" {
		model.Type = types.StringValue(networkType)
	}
	managedType := anyString(raw["networkType"])
	if managedType == "" {
		if model.Type.ValueString() == "ovn" {
			managedType = "vxlan"
		} else if model.Type.ValueString() == "bridge" {
			managedType = "bridge"
		}
	}
	if managedType != "" {
		model.NetworkType = types.StringValue(managedType)
	}
	config, _ := raw["config"].(map[string]any)
	if config != nil {
		converted := make(map[string]string, len(config))
		for key, value := range config {
			converted[key] = anyString(value)
		}
		model.Config, diagnostics = types.MapValueFrom(ctx, types.StringType, converted)
	}
	return diagnostics
}
func anyString(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}
