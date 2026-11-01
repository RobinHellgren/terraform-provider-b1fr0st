package provider

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = (*NetworkForwardResource)(nil)
var _ resource.ResourceWithConfigure = (*NetworkForwardResource)(nil)

type NetworkForwardResource struct{ client *APIClient }

type NetworkForwardModel struct {
	ID            types.String `tfsdk:"id"`
	ProjectID     types.String `tfsdk:"project_id"`
	NetworkName   types.String `tfsdk:"network_name"`
	ListenAddress types.String `tfsdk:"listen_address"`
	ListenPort    types.Int64  `tfsdk:"listen_port"`
	TargetAddress types.String `tfsdk:"target_address"`
	TargetPort    types.Int64  `tfsdk:"target_port"`
	Protocol      types.String `tfsdk:"protocol"`
	Description   types.String `tfsdk:"description"`
}

func NewNetworkForwardResource() resource.Resource { return &NetworkForwardResource{} }

func (r *NetworkForwardResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_forward"
}

func (r *NetworkForwardResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	stringReplace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	intReplace := []planmodifier.Int64{int64planmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":             schema.StringAttribute{Computed: true, Description: "Network name and listen address joined by a slash."},
		"project_id":     schema.StringAttribute{Required: true, PlanModifiers: stringReplace, Description: "Project UUID that owns the network."},
		"network_name":   schema.StringAttribute{Required: true, PlanModifiers: stringReplace, Description: "Managed Incus bridge or OVN network."},
		"listen_address": schema.StringAttribute{Required: true, PlanModifiers: stringReplace, Description: "Unique address available on the network uplink. OVN requires this address to be routed by its uplink."},
		"listen_port":    schema.Int64Attribute{Required: true, PlanModifiers: intReplace, Description: "TCP or UDP port exposed on the listen address."},
		"target_address": schema.StringAttribute{Required: true, PlanModifiers: stringReplace, Description: "IPv4 address of a workload attached to the selected network."},
		"target_port":    schema.Int64Attribute{Required: true, PlanModifiers: intReplace, Description: "Port on the target workload."},
		"protocol":       schema.StringAttribute{Required: true, PlanModifiers: stringReplace, Description: "tcp or udp."},
		"description":    schema.StringAttribute{Optional: true, PlanModifiers: stringReplace, Description: "Description shown by Incus."},
	}}
}

func (r *NetworkForwardResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected *APIClient, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *NetworkForwardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NetworkForwardModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := map[string]any{
		"listenAddress": plan.ListenAddress.ValueString(), "listenPort": plan.ListenPort.ValueInt64(),
		"targetAddress": plan.TargetAddress.ValueString(), "targetPort": plan.TargetPort.ValueInt64(),
		"protocol": plan.Protocol.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		input["description"] = plan.Description.ValueString()
	}
	path := "/networks/" + url.PathEscape(plan.NetworkName.ValueString()) + "/forwards"
	if _, err := r.client.Request(ctx, "POST", path, input, nil); err != nil {
		resp.Diagnostics.AddError("Create network forward failed", err.Error())
		return
	}
	plan.ID = types.StringValue(plan.NetworkName.ValueString() + "/" + plan.ListenAddress.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NetworkForwardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NetworkForwardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	path := "/networks/" + url.PathEscape(state.NetworkName.ValueString()) + "/forwards/" + url.PathEscape(state.ListenAddress.ValueString())
	var out struct {
		Forward map[string]any `json:"forward"`
	}
	status, err := r.client.Request(ctx, "GET", path, nil, &out)
	if status == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read network forward failed", err.Error())
		return
	}
	if out.Forward != nil {
		state.ListenAddress = types.StringValue(anyString(out.Forward["listen_address"]))
		if config, ok := out.Forward["config"].(map[string]any); ok {
			if value := anyString(config["target_address"]); value != "" {
				state.TargetAddress = types.StringValue(value)
			}
		}
		if ports, ok := out.Forward["ports"].([]any); ok && len(ports) > 0 {
			if port, ok := ports[0].(map[string]any); ok {
				state.Protocol = types.StringValue(anyString(port["protocol"]))
				state.TargetAddress = types.StringValue(anyString(port["target_address"]))
				if value, err := strconv.ParseInt(anyString(port["listen_port"]), 10, 64); err == nil {
					state.ListenPort = types.Int64Value(value)
				}
				if value, err := strconv.ParseInt(anyString(port["target_port"]), 10, 64); err == nil {
					state.TargetPort = types.Int64Value(value)
				}
				state.Description = types.StringValue(anyString(port["description"]))
			}
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NetworkForwardResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
}

func (r *NetworkForwardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NetworkForwardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	path := "/networks/" + url.PathEscape(state.NetworkName.ValueString()) + "/forwards/" + url.PathEscape(state.ListenAddress.ValueString())
	status, err := r.client.Request(ctx, "DELETE", path, nil, nil)
	if err != nil && status != 404 {
		resp.Diagnostics.AddError("Delete network forward failed", err.Error())
	}
}
