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

var _ resource.Resource = (*DiskResource)(nil)
var _ resource.ResourceWithConfigure = (*DiskResource)(nil)
var _ resource.ResourceWithImportState = (*DiskResource)(nil)

type DiskResource struct{ client *APIClient }
type DiskModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
	SizeGiB   types.Int64  `tfsdk:"size_gib"`
	Pool      types.String `tfsdk:"pool"`
	Project   types.String `tfsdk:"project"`
}

func NewDiskResource() resource.Resource { return &DiskResource{} }
func (r *DiskResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_disk"
}
func (r *DiskResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":         schema.StringAttribute{Computed: true, Description: "Opaque platform disk identifier."},
		"project_id": schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, Description: "Project UUID that owns this disk."},
		"name":       schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, Description: "Unique disk name within the project."},
		"size_gib":   schema.Int64Attribute{Required: true, Description: "Disk capacity in GiB. The platform allows online growth and protects existing data by disallowing shrink operations."},
		"pool":       schema.StringAttribute{Computed: true, Description: "Backing storage pool."},
		"project":    schema.StringAttribute{Computed: true, Description: "Incus project name."},
	}}
}
func (r *DiskResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *DiskResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *DiskResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DiskModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out struct {
		Disk map[string]any `json:"disk"`
	}
	_, err := r.client.Request(ctx, "POST", "/disks", map[string]any{"projectId": plan.ProjectID.ValueString(), "name": plan.Name.ValueString(), "sizeGiB": plan.SizeGiB.ValueInt64()}, &out)
	if err != nil {
		resp.Diagnostics.AddError("Create disk failed", err.Error())
		return
	}
	resp.Diagnostics.Append(diskModelFromAPI(&plan, out.Disk)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *DiskResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DiskModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out struct {
		Disk map[string]any `json:"disk"`
	}
	status, err := r.client.Request(ctx, "GET", "/disks/"+pathID(state.ID.ValueString()), nil, &out)
	if status == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read disk failed", err.Error())
		return
	}
	resp.Diagnostics.Append(diskModelFromAPI(&state, out.Disk)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *DiskResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DiskModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out struct {
		Disk map[string]any `json:"disk"`
	}
	_, err := r.client.Request(ctx, "PUT", "/disks/"+pathID(plan.ID.ValueString()), map[string]any{"sizeGiB": plan.SizeGiB.ValueInt64()}, &out)
	if err != nil {
		resp.Diagnostics.AddError("Resize disk failed", err.Error())
		return
	}
	resp.Diagnostics.Append(diskModelFromAPI(&plan, out.Disk)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *DiskResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DiskModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	status, err := r.client.Request(ctx, "DELETE", "/disks/"+pathID(state.ID.ValueString()), nil, nil)
	if err != nil && status != 404 {
		resp.Diagnostics.AddError("Delete disk failed", err.Error())
	}
}
func diskModelFromAPI(model *DiskModel, raw map[string]any) (diagnostics diag.Diagnostics) {
	if value := anyString(raw["id"]); value != "" {
		model.ID = types.StringValue(value)
	}
	if value := anyString(raw["projectId"]); value != "" {
		model.ProjectID = types.StringValue(value)
	}
	if value := anyString(raw["name"]); value != "" {
		model.Name = types.StringValue(value)
	}
	if value, ok := raw["sizeGiB"]; ok {
		model.SizeGiB = types.Int64Value(int64Value(value))
	}
	if value := anyString(raw["pool"]); value != "" {
		model.Pool = types.StringValue(value)
	}
	if value := anyString(raw["project"]); value != "" {
		model.Project = types.StringValue(value)
	}
	return nil
}
func int64Value(value any) int64 {
	switch n := value.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case int:
		return int64(n)
	default:
		return 0
	}
}
